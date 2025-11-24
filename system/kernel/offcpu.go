package kernel

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	TargetPID string // 필수: 모니터링할 PID
}

type OffCPUData struct {
	Timestamp   string   `json:"timestamp"`    // 수집 시각
	ProcessName string   `json:"process_name"` // 프로세스 이름
	PID 		string 	 `json:"pid"`		   // 프로세스 ID
	TID         string   `json:"tid"`          // Thread ID
	DurationUs  int      `json:"duration_us"`  // 대기 시간 (마이크로초)
	Stack       []string `json:"stack"`        // 스택 트레이스
}

func Start(ctx context.Context, wg *sync.WaitGroup, outChan chan<- OffCPUData, cfg Config) {
	defer wg.Done()

	if cfg.TargetPID == "" {
		log.Println("[Kernel] Target PID is missing for Off-CPU tracing.")
		return
	}
	cmdPath := "/usr/sbin/offcputime-bpfcc"
	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		cmdPath = "offcputime"
	}
	if _, err := exec.LookPath(cmdPath); err != nil {
		if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
			log.Printf("[Kernel] 'offcputime' tool not found. Please install bcc-tools.\n")
			return
		}
	}

	fmt.Printf("[Kernel] Started Off-CPU Monitor for PID: %s (Cmd: %s)\n", cfg.TargetPID, cmdPath)
	reProcess := regexp.MustCompile(`-\s+(.+)\s+\((\d+)\)`)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("[Kernel] Stopping Off-CPU Monitor.")
			return
		default:
			cmd := exec.CommandContext(ctx, cmdPath, "-p", cfg.TargetPID, "1")
			output, err := cmd.CombinedOutput()
			if err != nil {
				if ctx.Err() == nil {
					log.Printf("[Kernel] Execution error: %v\n", err)
				}
				time.Sleep(1 * time.Second)
				continue
			}

			nowStr := time.Now().Format("2006-01-02 15:04:05.000000")
			results := parseOutput(string(output), reProcess, nowStr, cfg.TargetPID)
			for _, data := range results {
				select {
				case outChan <- data:
				case <-ctx.Done():
					return
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func parseOutput(output string, reProcess *regexp.Regexp, timestamp string, pid string) []OffCPUData {
	var results []OffCPUData
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentStack []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "Tracing") {
			continue
		}

		if strings.HasPrefix(line, "-") && strings.Contains(line, "(") {
			matches := reProcess.FindStringSubmatch(line)
			if len(matches) == 3 {
				procName := strings.TrimSpace(matches[1])
				tid := matches[2]

				if scanner.Scan() {
					timeLine := strings.TrimSpace(scanner.Text())
					duration, err := strconv.Atoi(timeLine)
					if err == nil {
						results = append(results, OffCPUData{
							Timestamp:   timestamp,
							ProcessName: procName,
							PID:		 pid,
							TID:         tid,
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
