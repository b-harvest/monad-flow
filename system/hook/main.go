package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	TargetBinary = "/usr/local/bin/monad-node"
	TargetOffset = "0x96C720"
)

// JSON 출력을 위한 최상위 구조체
type TraceLog struct {
	EventType string      `json:"event_type"`
	Timestamp string      `json:"timestamp"`
	PID       string      `json:"pid"`
	TID       string      `json:"tid"`
	Data      interface{} `json:"data"`
}

type EnterData struct {
	TargetRaw   string   `json:"func_raw"`
	TargetClean string   `json:"func_clean"`
	CallerRaw   string   `json:"caller_raw"`
	CallerClean string   `json:"caller_clean"`
	Args        []string `json:"args_hex"`
}

type ExitData struct {
	DurationNs  string `json:"duration_ns"`
	BackToRaw   string `json:"back_to_raw"`
	BackToClean string `json:"back_to_clean"`
	ReturnValue string `json:"return_value"`
}

func main() {
	bpftraceScript := fmt.Sprintf(`
    uprobe:%s:%s {
        @start[tid] = nsecs;
        printf("E|%%d|%%d|%%s|%%s|%%s|0x%%lx|0x%%lx|0x%%lx|0x%%lx|0x%%lx|0x%%lx\n", 
            pid, tid, strftime("%%H:%%M:%%S.%%f", nsecs), 
            usym(reg("ip")), usym(*reg("sp")),
            arg0, arg1, arg2, arg3, arg4, arg5);
    }

    uretprobe:%s:%s {
        if (@start[tid]) {
            $dur_ns = nsecs - @start[tid];
            printf("X|%%d|%%d|%%d|%%s|0x%%lx\n", 
                pid, tid, $dur_ns, usym(reg("ip")), retval);
            delete(@start[tid]);
        }
    }
    `, TargetBinary, TargetOffset, TargetBinary, TargetOffset)

	cmd := exec.Command("bpftrace", "-e", bpftraceScript)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Pipe error:", err)
		return
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println("Start error:", err)
		return
	}

	fmt.Fprintf(os.Stderr, "Tracing %s (Offset %s)...\n", TargetBinary, TargetOffset)
	fmt.Fprintln(os.Stderr, "----------------------------------------------------------------------------------")

	scanner := bufio.NewScanner(stdout)

	// [중요 변경] JSON 인코더 생성 및 HTML 이스케이프 비활성화
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false) // <--- 이 부분이 추가되었습니다!

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Attaching") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}

		eventType := parts[0]

		if eventType == "E" && len(parts) >= 12 {
			pid := parts[1]
			tid := parts[2]
			timeStr := parts[3]
			targetRaw := parts[4]
			callerRaw := parts[5]
			args := parts[6:12]

			logEntry := TraceLog{
				EventType: "ENTER",
				Timestamp: timeStr,
				PID:       pid,
				TID:       tid,
				Data: EnterData{
					TargetRaw:   targetRaw,
					TargetClean: cleanSymbol(targetRaw),
					CallerRaw:   callerRaw,
					CallerClean: cleanSymbol(callerRaw),
					Args:        args,
				},
			}
			encoder.Encode(logEntry)

		} else if eventType == "X" && len(parts) >= 6 {
			pid := parts[1]
			tid := parts[2]
			durNs := parts[3]
			backToRaw := parts[4]
			retVal := parts[5]

			logEntry := TraceLog{
				EventType: "EXIT",
				Timestamp: time.Now().Format("15:04:05.000000"),
				PID:       pid,
				TID:       tid,
				Data: ExitData{
					DurationNs:  durNs,
					BackToRaw:   backToRaw,
					BackToClean: cleanSymbol(backToRaw),
					ReturnValue: retVal,
				},
			}
			encoder.Encode(logEntry)
		}
	}
	cmd.Wait()
}

func cleanSymbol(sym string) string {
	s := sym
	s = strings.ReplaceAll(s, "_$LT$", "<")
	s = strings.ReplaceAll(s, "$LT$", "<")
	s = strings.ReplaceAll(s, "$GT$", ">")
	s = strings.ReplaceAll(s, "$u20$", " ")
	s = strings.ReplaceAll(s, "$C$", ",")
	s = strings.ReplaceAll(s, "..", "::")
	return s
}
