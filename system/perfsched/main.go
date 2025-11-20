package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// SchedStat: /proc/[pid]/schedstat 정보를 담는 구조체
type SchedStat struct {
	RunTime   uint64
	WaitTime  uint64
	RunCount  uint64
	Timestamp time.Time
}

// 스레드 정보를 담는 구조체
type ThreadInfo struct {
	TID      string
	Comm     string
	Current  SchedStat
	Previous SchedStat
}

// JSON 출력을 위한 구조체
type SchedLog struct {
	Timestamp   string  `json:"timestamp"`     // 로그 기록 시간
	MainPID     string  `json:"main_pid"`      // 모니터링 대상 메인 PID
	TID         string  `json:"tid"`           // 스레드 ID
	ThreadName  string  `json:"thread_name"`   // 스레드 이름
	WaitDeltaMs float64 `json:"wait_delta_ms"` // 대기 시간 변화량 (Latency)
	RunDeltaMs  float64 `json:"run_delta_ms"`  // 실행 시간 변화량
	CtxSwitches uint64  `json:"ctx_switches"`  // 컨텍스트 스위칭 횟수 변화량
}

// /proc/[tid]/schedstat 파일을 읽어서 파싱
func readSchedStat(path string) (SchedStat, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SchedStat{}, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return SchedStat{}, fmt.Errorf("invalid format")
	}

	runTime, _ := strconv.ParseUint(fields[0], 10, 64)
	waitTime, _ := strconv.ParseUint(fields[1], 10, 64)
	runCount, _ := strconv.ParseUint(fields[2], 10, 64)

	return SchedStat{
		RunTime:   runTime,
		WaitTime:  waitTime,
		RunCount:  runCount,
		Timestamp: time.Now(),
	}, nil
}

// /proc/[tid]/comm 파일을 읽어서 이름 파악
func readComm(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

// 수집된 데이터를 분석하여 JSON으로 출력
func printAnalysis(mainPid string, threads map[string]*ThreadInfo) {
	// JSON 인코더 설정 (매번 생성하거나, 외부에서 주입받아도 됨)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false) // <, > 등의 특수문자 이스케이프 방지

	nowStr := time.Now().Format("2006-01-02 15:04:05.000000")
	activeCount := 0

	for _, t := range threads {
		// 이전 데이터가 없으면(첫 실행) 스킵
		if t.Previous.RunCount == 0 && t.Previous.RunTime == 0 {
			continue
		}

		waitDelta := float64(t.Current.WaitTime-t.Previous.WaitTime) / 1_000_000.0
		runDelta := float64(t.Current.RunTime-t.Previous.RunTime) / 1_000_000.0
		switchDelta := t.Current.RunCount - t.Previous.RunCount

		// 변화가 없는 스레드는 출력하지 않음 (로그 양 조절)
		if switchDelta == 0 && waitDelta == 0 && runDelta == 0 {
			continue
		}

		logEntry := SchedLog{
			Timestamp:   nowStr,
			MainPID:     mainPid,
			TID:         t.TID,
			ThreadName:  t.Comm,
			WaitDeltaMs: waitDelta,
			RunDeltaMs:  runDelta,
			CtxSwitches: switchDelta,
		}

		if err := encoder.Encode(logEntry); err != nil {
			fmt.Fprintf(os.Stderr, "JSON Encoding Error: %v\n", err)
		}
		activeCount++
	}

	// 데이터가 발생했음을 stderr로 알림 (선택 사항)
	if activeCount > 0 {
		fmt.Fprintf(os.Stderr, "--- [Time: %s] %d개 스레드 활동 감지됨 ---\n",
			time.Now().Format("15:04:05"), activeCount)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "사용법: sudo go run . <PID>")
		os.Exit(1)
	}
	mainPid := os.Args[1]

	threads := make(map[string]*ThreadInfo)

	fmt.Fprintf(os.Stderr, "--- [PID %s] 스케줄러 모니터링 시작 (Ctrl+C로 종료) ---\n", mainPid)

	for {
		taskPath := filepath.Join("/proc", mainPid, "task")
		entries, err := os.ReadDir(taskPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading tasks: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// 현재 존재하는 스레드 목록 확인을 위한 맵
		currentTIDs := make(map[string]bool)

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
				threads[tid] = &ThreadInfo{
					TID:  tid,
					Comm: readComm(commPath),
				}
			}

			t := threads[tid]
			// 값을 갱신 (Previous <- Current)
			t.Previous = t.Current
			t.Current = currentStat
		}

		// 사라진 스레드 정리 (메모리 누수 방지)
		for tid := range threads {
			if !currentTIDs[tid] {
				delete(threads, tid)
			}
		}

		printAnalysis(mainPid, threads)
		time.Sleep(1 * time.Second)
	}
}
