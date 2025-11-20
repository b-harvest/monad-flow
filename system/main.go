package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	Source string      `json:"source"` // "BPF", "LOG", "OFF_CPU", "SCHED", "PERF", "TURBO"
	Data   interface{} `json:"data"`   // 각 모듈의 구조체
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using OS env vars.")
	}

	pid := os.Getenv("TARGET_PID")
	services := strings.Split(os.Getenv("TARGET_SERVICES"), ",") // 예: "monad-node.service,monad-bft.service"
	binaryPath := os.Getenv("TARGET_BINARY")
	offset := os.Getenv("TARGET_OFFSET")

	socketURL := os.Getenv("BACKEND_URL")
	if socketURL == "" {
		socketURL = "http://127.0.0.1:3000"
	}

	fmt.Println("==========================================")
	fmt.Println("All-in-One System Monitor Starting...")
	fmt.Printf("Target PID: %s\n", pid)
	fmt.Printf("Socket Server: %s\n", socketURL)
	fmt.Println("==========================================")

	mainDataChan := make(chan DataPacket, 1000)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go runWebSocketManager(ctx, &wg, mainDataChan, socketURL)

	hookChan := make(chan hook.TraceLog, 100)
	wg.Add(1)
	go hook.Start(ctx, &wg, hookChan, hook.Config{TargetPID: pid, BinaryPath: binaryPath, Offset: offset})
	go bridge(ctx, hookChan, mainDataChan, "BPF_TRACE")

	journalChan := make(chan journalctl.LogEntry, 100)
	wg.Add(1)
	go journalctl.Start(ctx, &wg, journalChan, journalctl.Config{Services: services})
	go bridge(ctx, journalChan, mainDataChan, "SYSTEM_LOG")

	offCpuChan := make(chan kernel.OffCPUData, 100)
	wg.Add(1)
	go kernel.Start(ctx, &wg, offCpuChan, kernel.Config{TargetPID: pid})
	go bridge(ctx, offCpuChan, mainDataChan, "OFF_CPU")

	schedChan := make(chan scheduler.SchedLog, 100)
	wg.Add(1)
	go scheduler.Start(ctx, &wg, schedChan, scheduler.Config{TargetPID: pid})
	go bridge(ctx, schedChan, mainDataChan, "SCHEDULER")

	perfChan := make(chan perf.PerfLog, 50)
	wg.Add(1)
	go perf.Start(ctx, &wg, perfChan, perf.Config{TargetPID: pid})
	go bridge(ctx, perfChan, mainDataChan, "PERF_STAT")

	turboChan := make(chan turbostat.TurbostatMetric, 50)
	wg.Add(1)
	go turbostat.Start(ctx, &wg, turboChan, turbostat.Config{TargetPID: pid})
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

	client.On("connect", func(data ...any) {
		log.Printf("[WebSocket] Connected to Server! (ID: %v)", client.Id())
	})

	client.On("disconnect", func(data ...any) {
		log.Println("[WebSocket] Disconnected from Server.")
	})

	client.On("connect_error", func(err ...any) {
		log.Printf("[WebSocket] Connection Error: %v", err[0])
	})

	fmt.Println("WebSocket Sender Started. Forwarding data...")
	for {
		select {
		case packet, ok := <-input:
			if !ok {
				return
			}

			if client.Connected() {
				client.Emit(packet.Source, packet.Data)
			} else {
				fmt.Printf("Drop [%s] (Not Connected)\n", packet.Source)
				jsonData, err := json.Marshal(packet.Data)
				if err == nil {
					fmt.Printf("   └─ Data: %s\n", string(jsonData))
				} else {
					fmt.Printf("   └─ Data (Raw): %+v\n", packet.Data)
				}
			}

		case <-ctx.Done():
			fmt.Println("Stopping WebSocket Manager...")
			return
		}
	}
}
