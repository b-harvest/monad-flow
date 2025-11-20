package hook

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	FallbackBinary = "/usr/local/bin/monad-node"
	FallbackOffset = "0x96C720"
)

type Config struct {
	TargetPID  string // 모니터링할 PID (비어있으면 전체)
	BinaryPath string // 바이너리 경로 (예: /usr/bin/node)
	Offset     string // 오프셋 (예: 0x96C720)
}

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

func Start(ctx context.Context, wg *sync.WaitGroup, outChan chan<- TraceLog, cfg Config) {
	defer wg.Done()

	targetBin := cfg.BinaryPath
	if targetBin == "" {
		targetBin = FallbackBinary
	}

	targetOffset := cfg.Offset
	if targetOffset == "" {
		targetOffset = FallbackOffset
	}

	pidFilter := ""
	if cfg.TargetPID != "" {
		pidFilter = fmt.Sprintf(" /pid == %s/ ", cfg.TargetPID)
	}

	bpftraceScript := fmt.Sprintf(`
    uprobe:%s:%s %s {
        @start[tid] = nsecs;
        printf("E|%%d|%%d|%%s|%%s|%%s|0x%%lx|0x%%lx|0x%%lx|0x%%lx|0x%%lx|0x%%lx\n", 
            pid, tid, strftime("%%Y-%%m-%%d %%H:%%M:%%S.%%f", nsecs), 
            usym(reg("ip")), usym(*reg("sp")),
            arg0, arg1, arg2, arg3, arg4, arg5);
    }

    uretprobe:%s:%s %s {
        if (@start[tid]) {
            $dur_ns = nsecs - @start[tid];
            printf("X|%%d|%%d|%%d|%%s|0x%%lx\n", 
                pid, tid, $dur_ns, usym(reg("ip")), retval);
            delete(@start[tid]);
        }
    }
    `, targetBin, targetOffset, pidFilter, targetBin, targetOffset, pidFilter)

	cmd := exec.CommandContext(ctx, "bpftrace", "-e", bpftraceScript)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[Hook] Pipe error: %v\n", err)
		return
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("[Hook] Start error (Check sudo permissions): %v\n", err)
		return
	}

	fmt.Printf("[Hook] Started tracing %s (Offset: %s, PID Filter: %s)\n", targetBin, targetOffset, cfg.TargetPID)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "Attaching") || line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}

		eventType := parts[0]
		var logEntry TraceLog
		isValid := false

		if eventType == "E" && len(parts) >= 12 {
			logEntry = TraceLog{
				EventType: "ENTER",
				Timestamp: parts[3],
				PID:       parts[1],
				TID:       parts[2],
				Data: EnterData{
					TargetRaw:   parts[4],
					TargetClean: cleanSymbol(parts[4]),
					CallerRaw:   parts[5],
					CallerClean: cleanSymbol(parts[5]),
					Args:        parts[6:12],
				},
			}
			isValid = true
		} else if eventType == "X" && len(parts) >= 6 {
			logEntry = TraceLog{
				EventType: "EXIT",
				Timestamp: time.Now().Format("2006-01-02 15:04:05.000000"),
				PID:       parts[1],
				TID:       parts[2],
				Data: ExitData{
					DurationNs:  parts[3],
					BackToRaw:   parts[4],
					BackToClean: cleanSymbol(parts[4]),
					ReturnValue: parts[5],
				},
			}
			isValid = true
		}

		if isValid {
			select {
			case outChan <- logEntry:
			case <-ctx.Done():
				return
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == nil {
			log.Printf("[Hook] Process exited with error: %v\n", err)
		}
	}
	fmt.Println("[Hook] Service stopped.")
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
