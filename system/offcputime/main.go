package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type OffCPUData struct {
	ProcessName string
	PID         string
	DurationUs  int
	Stack       []string
}

func printFormattedData(data []OffCPUData) {
	sort.Slice(data, func(i, j int) bool {
		return data[i].DurationUs > data[j].DurationUs
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "TID\tPROCESS\tLATENCY(ms)\tSTACK TRACE (Call Chain)")
	fmt.Fprintln(w, "---\t-------\t-----------\t------------------------")

	for _, d := range data {
		ms := float64(d.DurationUs) / 1000.0

		stackStr := strings.Join(d.Stack, " -> ")
		fmt.Fprintf(w, "%s\t%s\t%.3f ms\t[ %s ]\n", d.PID, d.ProcessName, ms, stackStr)
	}

	w.Flush()
	fmt.Printf("--- 총 %d건의 대기 이벤트 감지됨 ---\n", len(data))
}

func parseOutput(output string) []OffCPUData {
	var results []OffCPUData
	scanner := bufio.NewScanner(strings.NewReader(output))

	reProcess := regexp.MustCompile(`-\s+(.+)\s+\((\d+)\)`)

	var currentStack []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Tracing") {
			continue
		}
		if strings.HasPrefix(line, "-") && strings.Contains(line, "(") {
			matches := reProcess.FindStringSubmatch(line)
			if len(matches) == 3 {
				procName := strings.TrimSpace(matches[1])
				pid := matches[2]

				if scanner.Scan() {
					timeLine := strings.TrimSpace(scanner.Text())
					duration, err := strconv.Atoi(timeLine)
					if err == nil {
						results = append(results, OffCPUData{
							ProcessName: procName,
							PID:         pid,
							DurationUs:  duration,
							Stack:       append([]string{}, currentStack...),
						})
					}
				}
				currentStack = []string{}
			}
		} else {
			if _, err := strconv.Atoi(line); err != nil {
				currentStack = append(currentStack, line)
			}
		}
	}
	return results
}

func runCommand(cmdPath string, pid string, duration string) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf(">>> [측정 중...] PID: %s, Duration: %s초 (Time: %s) <<<\n",
		pid, duration, time.Now().Format("15:04:05"))
	fmt.Println(strings.Repeat("=", 80))

	cmd := exec.Command(cmdPath, "-p", pid, duration)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("--- [에러]: 실행 실패: %v\n", err)
		return
	}

	data := parseOutput(string(output))
	printFormattedData(data)
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("ERROR: 이 프로그램은 sudo 권한으로 실행해야 합니다.")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("사용법: sudo go run . <PID>")
		os.Exit(1)
	}
	pid := os.Args[1]

	cmdPath := "/usr/sbin/offcputime-bpfcc"

	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		cmdPath = "offcputime"
	}

	fmt.Printf("--- [PID %s] 모니터링 시작 (Ctrl+C로 종료) ---\n", pid)

	for {
		runCommand(cmdPath, pid, "1")
		time.Sleep(100 * time.Millisecond)
	}
}
