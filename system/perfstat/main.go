package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// 개별 메트릭 정보를 더 세분화하여 정의
type PerfMetric struct {
	Event     string `json:"event"`                // 예: cycles
	Value     string `json:"value"`                // 측정값 (예: 67202022)
	Unit      string `json:"unit,omitempty"`       // 단위 (예: msec) - 없는 경우 생략
	MetricVal string `json:"metric_val,omitempty"` // 주요 지표 (예: 3.597 GHz, 0.037 CPUs utilized)
	RunPct    string `json:"run_pct,omitempty"`    // 실행 비율 (예: 91.90%) - 괄호 안의 값
}

// 한 주기(Interval)의 데이터를 묶을 구조체
type PerfLog struct {
	Timestamp     string       `json:"timestamp"`      // 수집된 시스템 시간
	PerfTimestamp string       `json:"perf_timestamp"` // perf 실행 후 경과 시간 (예: 0.500506)
	PID           string       `json:"pid"`
	Metrics       []PerfMetric `json:"metrics"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("사용법: sudo go run . <PID>")
	}
	pid := os.Args[1]

	// perf stat 실행 (-I 500: 500ms 간격)
	cmd := exec.Command("perf", "stat", "-d", "-p", pid, "-I", "500")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	log.Printf("... PID %s perf 모니터링 시작 ...\n", pid)

	scanner := bufio.NewScanner(stderr)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)

	var currentLog *PerfLog

	for scanner.Scan() {
		line := scanner.Text()

		// 주석이나 빈 줄 무시
		if strings.HasPrefix(strings.TrimSpace(line), "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parsedMetric, perfTs, err := parsePerfLine(line)
		if err != nil {
			continue
		}

		// 타임스탬프가 바뀌면 이전 로그 덤프 후 초기화
		if currentLog != nil && currentLog.PerfTimestamp != perfTs {
			encoder.Encode(currentLog)
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
		encoder.Encode(currentLog)
	}

	cmd.Wait()
}

// 정규식: 괄호로 된 퍼센트 수치 추출용 -> (91.90%)
var reRunPct = regexp.MustCompile(`\(\s*([\d\.]+%)\s*\)$`)

func parsePerfLine(line string) (PerfMetric, string, error) {
	// 1. # 기준으로 좌(데이터) 우(설명) 분리
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

	// [0]: Timestamp
	timestamp := fields[0]

	p := PerfMetric{}

	// 2. 우측(설명) 파싱: MetricVal과 RunPct 분리
	// 예: "3.597 GHz                               (91.90%)"
	// MetricVal -> "3.597 GHz"
	// RunPct    -> "91.90%"
	if rightSide != "" {
		// 정규식으로 끝부분에 (xx%)가 있는지 확인
		match := reRunPct.FindStringSubmatch(rightSide)
		if len(match) > 1 {
			p.RunPct = match[1] // 괄호 안의 값만 추출
			// 원본 문자열에서 (xx%) 부분을 제거하고 나머지를 MetricVal로 설정
			p.MetricVal = strings.TrimSpace(strings.Replace(rightSide, match[0], "", 1))
		} else {
			// 퍼센트가 없으면 전체를 MetricVal로
			p.MetricVal = rightSide
		}
	}

	// 3. 좌측(데이터) 파싱
	// <not supported> 처리
	if strings.Contains(leftSide, "<not supported>") {
		p.Value = "Not Supported"
		if len(fields) > 1 {
			// timestamp 뒤에 나오는 것들을 합쳐서 이벤트명으로 추정
			p.Event = strings.Join(fields[1:], " ")
			// "<not supported>" 문자열 제거하고 정리
			p.Event = strings.TrimSpace(strings.ReplaceAll(p.Event, "<not supported>", ""))
		}
	} else {
		// 일반 포맷: [Timestamp] [Value] [Unit?] [Event]
		// fields[0]: Timestamp
		// fields[1]: Value
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
