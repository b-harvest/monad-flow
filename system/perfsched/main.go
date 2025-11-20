package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
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

// 수집된 데이터를 분석하여 출력
func printAnalysis(threads map[string]*ThreadInfo) {
	type Result struct {
		TID         string
		Name        string
		WaitDeltaMs float64
		RunDeltaMs  float64
		CtxSwitches uint64
	}

	var results []Result

	for _, t := range threads {
		if t.Previous.RunCount == 0 && t.Previous.RunTime == 0 {
			continue
		}
		waitDelta := float64(t.Current.WaitTime-t.Previous.WaitTime) / 1_000_000.0
		runDelta := float64(t.Current.RunTime-t.Previous.RunTime) / 1_000_000.0
		switchDelta := t.Current.RunCount - t.Previous.RunCount

		if switchDelta == 0 && waitDelta == 0 {
			continue
		}

		results = append(results, Result{
			TID:         t.TID,
			Name:        t.Comm,
			WaitDeltaMs: waitDelta,
			RunDeltaMs:  runDelta,
			CtxSwitches: switchDelta,
		})
	}

	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].WaitDeltaMs < results[j].WaitDeltaMs {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	fmt.Printf("\n>>> [스케줄러 지연(Latency) 실시간 분석] Time: %s <<<\n", time.Now().Format("15:04:05"))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "TID\tTHREAD NAME\tWAIT(Latency)\tCPU RUN\tSWITCHES")
	fmt.Fprintln(w, "---\t-----------\t-------------\t-------\t--------")

	count := 0
	for _, r := range results {
		fmt.Fprintf(w, "%s\t%s\t%.2f ms\t%.2f ms\t%d\n",
			r.TID, r.Name, r.WaitDeltaMs, r.RunDeltaMs, r.CtxSwitches)

		count++
		if count >= 20 {
			break
		}
	}
	w.Flush()
	fmt.Printf("--- 총 %d개 스레드 모니터링 중 ---\n", len(results))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("사용법: sudo go run . <PID>")
		os.Exit(1)
	}
	mainPid := os.Args[1]

	threads := make(map[string]*ThreadInfo)

	fmt.Printf("--- [PID %s] 모니터링 시작 (Ctrl+C로 종료) ---\n", mainPid)

	for {
		taskPath := filepath.Join("/proc", mainPid, "task")
		entries, err := os.ReadDir(taskPath)
		if err != nil {
			fmt.Printf("Error reading tasks: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, entry := range entries {
			tid := entry.Name()
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
			t.Previous = t.Current
			t.Current = currentStat
		}
		printAnalysis(threads)
		time.Sleep(1 * time.Second)
	}
}
