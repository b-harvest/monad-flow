package scheduler

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	TargetPID string // 필수: 모니터링할 PID
}

type SchedLog struct {
	Timestamp   string  `json:"timestamp"`     // 로그 기록 시간
	MainPID     string  `json:"main_pid"`      // 모니터링 대상 메인 PID
	TID         string  `json:"tid"`           // 스레드 ID
	ThreadName  string  `json:"thread_name"`   // 스레드 이름
	WaitDeltaMs float64 `json:"wait_delta_ms"` // 대기 시간 변화량 (Latency)
	RunDeltaMs  float64 `json:"run_delta_ms"`  // 실행 시간 변화량
	CtxSwitches uint64  `json:"ctx_switches"`  // 컨텍스트 스위칭 횟수 변화량
}

type schedStat struct {
	RunTime   uint64
	WaitTime  uint64
	RunCount  uint64
	Timestamp time.Time
}

type threadInfo struct {
	TID      string
	Comm     string
	Current  schedStat
	Previous schedStat
}

func Start(ctx context.Context, wg *sync.WaitGroup, outChan chan<- SchedLog, cfg Config) {
	defer wg.Done()

	if cfg.TargetPID == "" {
		log.Println("[Scheduler] Target PID is missing.")
		return
	}

	mainPid := cfg.TargetPID
	threads := make(map[string]*threadInfo)

	fmt.Printf("[Scheduler] Started Scheduler Monitor for PID: %s\n", mainPid)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[Scheduler] Stopping Scheduler Monitor.")
			return
		case <-ticker.C:
			taskPath := filepath.Join("/proc", mainPid, "task")
			entries, err := os.ReadDir(taskPath)
			if err != nil {
				continue
			}

			currentTIDs := make(map[string]bool)
			activeCount := 0

			nowStr := time.Now().Format("2006-01-02 15:04:05.000000")

			for _, entry := range entries {
				tid := entry.Name()
				currentTIDs[tid] = true

				statPath := filepath.Join(taskPath, tid, "schedstat")
				commPath := filepath.Join(taskPath, tid, "comm")

				currentStat, err := readSchedStat(statPath)
				if err != nil {
					continue
				}

				if _, exists := threads[tid]; !exists {
					threads[tid] = &threadInfo{
						TID:  tid,
						Comm: readComm(commPath),
					}
				}

				t := threads[tid]
				t.Previous = t.Current
				t.Current = currentStat

				if t.Previous.RunCount == 0 && t.Previous.RunTime == 0 {
					continue
				}

				waitDelta := float64(t.Current.WaitTime-t.Previous.WaitTime) / 1_000_000.0
				runDelta := float64(t.Current.RunTime-t.Previous.RunTime) / 1_000_000.0
				switchDelta := t.Current.RunCount - t.Previous.RunCount

				if switchDelta > 0 || waitDelta > 0 || runDelta > 0 {
					logEntry := SchedLog{
						Timestamp:   nowStr,
						MainPID:     mainPid,
						TID:         t.TID,
						ThreadName:  t.Comm,
						WaitDeltaMs: waitDelta,
						RunDeltaMs:  runDelta,
						CtxSwitches: switchDelta,
					}

					select {
					case outChan <- logEntry:
						activeCount++
					case <-ctx.Done():
						return
					}
				}
			}

			for tid := range threads {
				if !currentTIDs[tid] {
					delete(threads, tid)
				}
			}
		}
	}
}

func readSchedStat(path string) (schedStat, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return schedStat{}, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return schedStat{}, fmt.Errorf("invalid format")
	}

	runTime, _ := strconv.ParseUint(fields[0], 10, 64)
	waitTime, _ := strconv.ParseUint(fields[1], 10, 64)
	runCount, _ := strconv.ParseUint(fields[2], 10, 64)

	return schedStat{
		RunTime:   runTime,
		WaitTime:  waitTime,
		RunCount:  runCount,
		Timestamp: time.Now(),
	}, nil
}

func readComm(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}
