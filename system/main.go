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

	fmt.Println("==========================================")
	fmt.Println("🚀 All-in-One System Monitor Starting...")
	fmt.Printf("🎯 Target PID: %s\n", pid)
	fmt.Printf("📦 Services: %v\n", services)
	fmt.Println("==========================================")

	mainDataChan := make(chan DataPacket, 1000)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go runWebSocketManager(ctx, &wg, mainDataChan)

	hookChan := make(chan hook.TraceLog, 100)
	wg.Add(1)
	go hook.Start(ctx, &wg, hookChan, hook.Config{
		TargetPID:  pid,
		BinaryPath: binaryPath,
		Offset:     offset,
	})
	go bridge(ctx, hookChan, mainDataChan, "BPF_TRACE")

	journalChan := make(chan journalctl.LogEntry, 100)
	wg.Add(1)
	go journalctl.Start(ctx, &wg, journalChan, journalctl.Config{
		Services: services,
	})
	go bridge(ctx, journalChan, mainDataChan, "SYSTEM_LOG")

	offCpuChan := make(chan kernel.OffCPUData, 100)
	wg.Add(1)
	go kernel.Start(ctx, &wg, offCpuChan, kernel.Config{
		TargetPID: pid,
	})
	go bridge(ctx, offCpuChan, mainDataChan, "OFF_CPU")

	schedChan := make(chan scheduler.SchedLog, 100)
	wg.Add(1)
	go scheduler.Start(ctx, &wg, schedChan, scheduler.Config{
		TargetPID: pid,
	})
	go bridge(ctx, schedChan, mainDataChan, "SCHEDULER")

	perfChan := make(chan perf.PerfLog, 50)
	wg.Add(1)
	go perf.Start(ctx, &wg, perfChan, perf.Config{
		TargetPID: pid,
	})
	go bridge(ctx, perfChan, mainDataChan, "PERF_STAT")

	turboChan := make(chan turbostat.TurbostatMetric, 50)
	wg.Add(1)
	go turbostat.Start(ctx, &wg, turboChan, turbostat.Config{
		TargetPID: pid,
	})
	go bridge(ctx, turboChan, mainDataChan, "TURBO_STAT")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	fmt.Println("\n🛑 Shutting down system...")

	cancel()
	wg.Wait()
	close(mainDataChan)

	fmt.Println("✅ Monitor system exited cleanly.")
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

// runWebSocketManager: 최종적으로 데이터를 받아 WebSocket으로 전송하는 역할 (현재는 화면 출력)
func runWebSocketManager(ctx context.Context, wg *sync.WaitGroup, input <-chan DataPacket) {
	defer wg.Done()
	fmt.Println("📡 WebSocket Manager Started (Waiting for data...)")

	// JSON 인코더 (화면 출력용, 나중에 ws.WriteJSON으로 대체)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)

	for {
		select {
		case packet, ok := <-input:
			if !ok {
				return
			}

			// [TODO] 여기에 WebSocket 전송 로직이 들어갑니다.
			// 예: err := wsConn.WriteJSON(packet)

			// 현재는 시뮬레이션을 위해 콘솔에 출력
			// 데이터가 너무 빠르면 보기가 힘드니 간단하게 출력 포맷팅
			// fmt.Printf("[WS SEND] Source: %-10s | Data: %+v\n", packet.Source, packet.Data)

			// 또는 전체 JSON 덤프 (디버깅용)
			_ = encoder.Encode(packet)

		case <-ctx.Done():
			fmt.Println("📡 WebSocket Manager Stopped.")
			return
		}
	}
}
