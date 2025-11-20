package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// JSON 출력을 위한 구조체 정의
type OffCPUData struct {
	Timestamp   string   `json:"timestamp"`    // 언제 파싱(수집)되었는지
	ProcessName string   `json:"process_name"` // 프로세스 이름
	TID         string   `json:"tid"`          // [변경] PID -> TID (실제로는 Thread ID임)
	DurationUs  int      `json:"duration_us"`  // 대기 시간 (마이크로초)
	Stack       []string `json:"stack"`        // 스택 트레이스 배열
}

// 파싱 로직
func parseOutput(output string) []OffCPUData {
	var results []OffCPUData
	scanner := bufio.NewScanner(strings.NewReader(output))

	// 예: "- monad-node (12345)" 패턴 매칭
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

		// 프로세스 정보 라인 감지
		if strings.HasPrefix(line, "-") && strings.Contains(line, "(") {
			matches := reProcess.FindStringSubmatch(line)
			if len(matches) == 3 {
				procName := strings.TrimSpace(matches[1])
				tid := matches[2] // [변경] 여기서 추출한 숫자는 TID입니다.

				// 다음 줄에 있는 Duration(시간) 읽기
				if scanner.Scan() {
					timeLine := strings.TrimSpace(scanner.Text())
					duration, err := strconv.Atoi(timeLine)
					if err == nil {
						results = append(results, OffCPUData{
							// Timestamp는 출력 직전에 넣습니다.
							ProcessName: procName,
							TID:         tid, // [변경] 구조체 필드 TID에 할당
							DurationUs:  duration,
							Stack:       append([]string{}, currentStack...),
						})
					}
				}
				// 스택 초기화
				currentStack = []string{}
			}
		} else {
			// 숫자가 아닌 경우 스택의 일부로 간주
			if _, err := strconv.Atoi(line); err != nil {
				currentStack = append(currentStack, line)
			}
		}
	}
	return results
}

func runCommand(cmdPath string, targetPid string, duration string) {
	// 진행 상황 로그 (Stderr)
	fmt.Fprintf(os.Stderr, ">>> [측정 중...] Target PID: %s, Duration: %s초 (Time: %s)\n",
		targetPid, duration, time.Now().Format("15:04:05"))

	cmd := exec.Command(cmdPath, "-p", targetPid, duration)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Fprintf(os.Stderr, "--- [에러]: 실행 실패: %v\n", err)
		return
	}

	// 1. 파싱
	data := parseOutput(string(output))

	// 2. JSON 인코더 설정 (HTML 이스케이프 끄기)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)

	// 3. 타임스탬프 주입 및 JSON 출력
	nowStr := time.Now().Format("2006-01-02 15:04:05.000000")

	for _, d := range data {
		d.Timestamp = nowStr
		if err := encoder.Encode(d); err != nil {
			fmt.Fprintf(os.Stderr, "JSON Encoding Error: %v\n", err)
		}
	}

	// 결과 요약 (Stderr)
	if len(data) == 0 {
		fmt.Fprintf(os.Stderr, "--- 데이터 없음 ---\n")
	} else {
		fmt.Fprintf(os.Stderr, "--- %d건 JSON 출력 완료 ---\n", len(data))
	}
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "ERROR: 이 프로그램은 sudo 권한으로 실행해야 합니다.")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "사용법: sudo go run . <TARGET_PID>")
		os.Exit(1)
	}
	pid := os.Args[1] // 모니터링 대상 프로세스 ID

	cmdPath := "/usr/sbin/offcputime-bpfcc"

	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		cmdPath = "offcputime"
	}

	fmt.Fprintf(os.Stderr, "--- [Target PID %s] 모니터링 시작 (Ctrl+C로 종료) ---\n", pid)

	for {
		runCommand(cmdPath, pid, "1")
		time.Sleep(100 * time.Millisecond)
	}
}
