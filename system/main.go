package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"system/hook"
	"system/journalctl"
	"system/kernel"
	"system/perf"
	"system/scheduler"
	"system/turbostat"

	"github.com/joho/godotenv"
	"github.com/zishang520/socket.io/clients/socket/v3"
)

type DataPacket struct {
	Source string      `json:"source"`
	Data   interface{} `json:"data"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using OS env vars.")
	}

	rawServices := os.Getenv("TARGET_SERVICES") // 예: "monad-node.service,nginx.service"
	var targetServices []string
	if rawServices != "" {
		parts := strings.Split(rawServices, ",")
		for _, s := range parts {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				targetServices = append(targetServices, trimmed)
			}
		}
	}

	var targetPIDs []string
	if len(targetServices) > 0 {
		var err error
		targetPIDs, err = getPIDsFromServices(targetServices)
		if err != nil {
			log.Printf("[Main] Error extracting PIDs from services: %v", err)
		}
	}

	targetPIDs = uniqueStrings(targetPIDs)

	// tracePid := os.Getenv("TARGET_TRACE_PID")
	// binaryPath := os.Getenv("TARGET_TRACE_BINARY")
	rawOffsets := os.Getenv("TARGET_TRACE_OFFSETS")
	var hookTargets []hook.HookTarget
	if rawOffsets != "" {
		pairs := strings.Split(rawOffsets, ",")
		for _, pair := range pairs {
			parts := strings.Split(pair, ":")
			if len(parts) == 2 {
				hookTargets = append(hookTargets, hook.HookTarget{
					Name:   strings.TrimSpace(parts[0]),
					Offset: strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	socketURL := os.Getenv("BACKEND_URL")
	if socketURL == "" {
		socketURL = "http://127.0.0.1:3000"
	}

	fmt.Println("==========================================")
	fmt.Println("All-in-One System Monitor Starting...")
	fmt.Printf("Target Services: %v\n", targetServices)
	fmt.Printf("Resolved PIDs:   %v\n", targetPIDs) // 추출된 PID 확인
	fmt.Printf("Socket Server:   %s\n", socketURL)
	fmt.Println("==========================================")

	if len(targetPIDs) == 0 {
		log.Println("[Main] Warning: No active PIDs found from services. Process monitoring will be skipped.")
	}

	mainDataChan := make(chan DataPacket, 1000)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go runWebSocketManager(ctx, &wg, mainDataChan, socketURL)

	offCpuChan := make(chan kernel.OffCPUData, 100)
	schedChan := make(chan scheduler.SchedLog, 100)
	perfChan := make(chan perf.PerfLog, 50)

	go bridge(ctx, offCpuChan, mainDataChan, "OFF_CPU")
	go bridge(ctx, schedChan, mainDataChan, "SCHEDULER")
	go bridge(ctx, perfChan, mainDataChan, "PERF_STAT")

	for _, pid := range targetPIDs {
		wg.Add(1)
		go kernel.Start(ctx, &wg, offCpuChan, kernel.Config{TargetPID: pid})

		wg.Add(1)
		go scheduler.Start(ctx, &wg, schedChan, scheduler.Config{TargetPID: pid})

		wg.Add(1)
		go perf.Start(ctx, &wg, perfChan, perf.Config{TargetPID: pid})

		fmt.Printf("[Main] Launched monitors for PID: %s\n", pid)
	}

	// hookChan := make(chan hook.TraceLog, 100)
	// wg.Add(1)
	// go hook.Start(ctx, &wg, hookChan, hook.Config{
	// 	TargetPID:  tracePid,
	// 	BinaryPath: binaryPath,
	// 	Targets:    hookTargets,
	// })
	// go bridge(ctx, hookChan, mainDataChan, "BPF_TRACE")

	journalChan := make(chan journalctl.LogEntry, 100)
	wg.Add(1)
	go journalctl.Start(ctx, &wg, journalChan, journalctl.Config{Services: targetServices})
	go bridge(ctx, journalChan, mainDataChan, "SYSTEM_LOG")

	turboChan := make(chan turbostat.TurbostatMetric, 50)
	wg.Add(1)
	go turbostat.Start(ctx, &wg, turboChan)
	go bridge(ctx, turboChan, mainDataChan, "TURBO_STAT")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	fmt.Println("\nShutting down system...")

	cancel()
	wg.Wait()
	close(mainDataChan)

	fmt.Println("Monitor system exited cleanly.")
}

func getPIDsFromServices(services []string) ([]string, error) {
	var pids []string
	for _, svc := range services {
		cmd := exec.Command("systemctl", "show", "-p", "MainPID", "--value", svc)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Printf("[Main] Failed to get PID for service %s: %v", svc, err)
			continue
		}
		pidStr := strings.TrimSpace(out.String())
		if pidStr == "" || pidStr == "0" {
			log.Printf("[Main] Service %s is not running or has no PID.", svc)
			continue
		}
		pids = append(pids, pidStr)
		log.Printf("[Main] Found PID %s for service %s", pidStr, svc)
	}
	return pids, nil
}

func uniqueStrings(input []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func bridge[T any](ctx context.Context, input <-chan T, output chan<- DataPacket, sourceName string) {
	for {
		select {
		case data, ok := <-input:
			if !ok {
				return
			}
			output <- DataPacket{
				Source: sourceName,
				Data:   data,
			}
		case <-ctx.Done():
			return
		}
	}
}

func runWebSocketManager(ctx context.Context, wg *sync.WaitGroup, input <-chan DataPacket, url string) {
	defer wg.Done()

	client, err := socket.Connect(url, nil)
	if err != nil {
		log.Printf("Initial Socket Connection Failed: %v (Will try to reconnect...)", err)
	}

	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	fmt.Println("WebSocket Sender Started. Forwarding data...")
	for {
		select {
		case packet, ok := <-input:
			if !ok {
				return
			}
			if client.Connected() {
				client.Emit(packet.Source, packet.Data)
			}
		case <-ctx.Done():
			return
		}
	}
}
