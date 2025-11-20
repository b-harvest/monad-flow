package perf

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Config struct {
	TargetPID string // 필수: 모니터링할 PID
}

type PerfMetric struct {
	Event     string `json:"event"`
	Value     string `json:"value"`
	Unit      string `json:"unit,omitempty"`
	MetricVal string `json:"metric_val,omitempty"`
	RunPct    string `json:"run_pct,omitempty"`
}

type PerfLog struct {
	Timestamp     string       `json:"timestamp"`
	PerfTimestamp string       `json:"perf_timestamp"`
	PID           string       `json:"pid"`
	Metrics       []PerfMetric `json:"metrics"`
}

var reRunPct = regexp.MustCompile(`\(\s*([\d\.]+%)\s*\)$`)

func Start(ctx context.Context, wg *sync.WaitGroup, outChan chan<- PerfLog, cfg Config) {
	defer wg.Done()

	if cfg.TargetPID == "" {
		log.Println("⚠️ [Perf] Target PID is missing.")
		return
	}
	pid := cfg.TargetPID

	if _, err := exec.LookPath("perf"); err != nil {
		log.Println("❌ [Perf] 'perf' command not found. Please install linux-tools.")
		return
	}

	cmd := exec.CommandContext(ctx, "perf", "stat", "-d", "-p", pid, "-I", "500")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("❌ [Perf] Pipe error: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("❌ [Perf] Start error (Need sudo?): %v\n", err)
		return
	}

	fmt.Printf("🟢 [Perf] Started Perf Monitor for PID: %s\n", pid)

	scanner := bufio.NewScanner(stderr)
	var currentLog *PerfLog

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parsedMetric, perfTs, err := parsePerfLine(line)
		if err != nil {
			continue
		}

		if currentLog != nil && currentLog.PerfTimestamp != perfTs {
			select {
			case outChan <- *currentLog:
			case <-ctx.Done():
				return
			}
			currentLog = nil
		}

		if currentLog == nil {
			currentLog = &PerfLog{
				Timestamp:     time.Now().Format("2006-01-02 15:04:05.000000"),
				PerfTimestamp: perfTs,
				PID:           pid,
				Metrics:       []PerfMetric{},
			}
		}
		currentLog.Metrics = append(currentLog.Metrics, parsedMetric)
	}

	if currentLog != nil {
		select {
		case outChan <- *currentLog:
		case <-ctx.Done():
		}
	}

	cmd.Wait()
	fmt.Println("🔴 [Perf] Service stopped.")
}

func parsePerfLine(line string) (PerfMetric, string, error) {
	parts := strings.SplitN(line, "#", 2)
	leftSide := strings.TrimSpace(parts[0])
	rightSide := ""
	if len(parts) > 1 {
		rightSide = strings.TrimSpace(parts[1])
	}

	fields := strings.Fields(leftSide)
	if len(fields) < 2 {
		return PerfMetric{}, "", fmt.Errorf("invalid line")
	}

	timestamp := fields[0]
	p := PerfMetric{}

	if rightSide != "" {
		match := reRunPct.FindStringSubmatch(rightSide)
		if len(match) > 1 {
			p.RunPct = match[1]
			p.MetricVal = strings.TrimSpace(strings.Replace(rightSide, match[0], "", 1))
		} else {
			p.MetricVal = rightSide
		}
	}

	if strings.Contains(leftSide, "<not supported>") {
		p.Value = "Not Supported"
		if len(fields) > 1 {
			p.Event = strings.Join(fields[1:], " ")
			p.Event = strings.TrimSpace(strings.ReplaceAll(p.Event, "<not supported>", ""))
		}
	} else {
		p.Value = fields[1]
		if len(fields) > 2 {
			if fields[2] == "msec" {
				p.Unit = "msec"
				if len(fields) > 3 {
					p.Event = strings.Join(fields[3:], " ")
				}
			} else {
				p.Event = strings.Join(fields[2:], " ")
			}
		}
	}

	return p, timestamp, nil
}
