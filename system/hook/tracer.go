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

	"golang.org/x/sys/unix"
)

func init() {
    var rLimit unix.Rlimit
    rLimit.Max = 1048576 
    rLimit.Cur = 1048576
    if err := unix.Setrlimit(unix.RLIMIT_NOFILE, &rLimit); err != nil {
        log.Printf("[Init] Failed to set RLIMIT_NOFILE: %v", err)
    }

    rLimit.Max = unix.RLIM_INFINITY
    rLimit.Cur = unix.RLIM_INFINITY
    if err := unix.Setrlimit(unix.RLIMIT_MEMLOCK, &rLimit); err != nil {
        log.Printf("[Init] Failed to set RLIMIT_MEMLOCK: %v", err)
    }
    
    log.Println("[Init] System resource limits raised successfully.")
}

const (
	FallbackBinary = "/usr/local/bin/monad-node"
)

type HookTarget struct {
	Name   string // 로그에서 구분할 함수 별칭 (예: "ProcessTx", "UpdateState")
	Offset string // 후킹할 메모리 오프셋 (예: "0x96C720")
}

type Config struct {
	TargetPID  string       // 모니터링할 PID (비어있으면 전체)
	BinaryPath string       // 바이너리 경로
	Targets    []HookTarget // 다중 후킹 대상 리스트
}

type TraceLog struct {
	EventType string      `json:"event_type"`
	FuncName  string      `json:"func_name"`
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

	if len(cfg.Targets) == 0 {
		log.Println("[Hook] No targets specified.")
		return
	}

	pidFilter := ""
	if cfg.TargetPID != "" {
		pidFilter = fmt.Sprintf(" /pid == %s/ ", cfg.TargetPID)
	}

	var bpfScript strings.Builder

	for i, target := range cfg.Targets {
		fnId := i

		entryBlock := fmt.Sprintf(`
		uprobe:%s:%s %s {
			@start[tid, %d] = nsecs;
			printf("E|%s|%%d|%%d|%%s|%%s|%%s|0x%%lx|0x%%lx|0x%%lx|0x%%lx|0x%%lx|0x%%lx\n", 
				pid, tid, strftime("%%Y-%%m-%%d %%H:%%M:%%S.%%f", nsecs), 
				usym(reg("ip")), usym(*reg("sp")),
				arg0, arg1, arg2, arg3, arg4, arg5);
		}
		`, targetBin, target.Offset, pidFilter, fnId, target.Name)

		bpfScript.WriteString(entryBlock)

		exitBlock := fmt.Sprintf(`
		uretprobe:%s:%s %s {
			$start_time = @start[tid, %d];
			if ($start_time) {
				$dur_ns = nsecs - $start_time;
				printf("X|%s|%%d|%%d|%%d|%%s|0x%%lx\n", 
					pid, tid, $dur_ns, usym(reg("ip")), retval);
				delete(@start[tid, %d]);
			}
		}
		`, targetBin, target.Offset, pidFilter, fnId, target.Name, fnId)

		bpfScript.WriteString(exitBlock)
	}

	cmd := exec.CommandContext(ctx, "bpftrace", "-e", bpfScript.String())

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

	fmt.Printf("[Hook] Started tracing %d targets on %s (PID Filter: %s)\n", len(cfg.Targets), targetBin, cfg.TargetPID)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "Attaching") || line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		eventType := parts[0]
		funcName := parts[1]
		var logEntry TraceLog
		isValid := false

		if eventType == "E" && len(parts) >= 13 {
			logEntry = TraceLog{
				EventType: "ENTER",
				FuncName:  funcName,
				PID:       parts[2],
				TID:       parts[3],
				Timestamp: parts[4],
				Data: EnterData{
					TargetRaw:   parts[5],
					TargetClean: cleanSymbol(parts[5]),
					CallerRaw:   parts[6],
					CallerClean: cleanSymbol(parts[6]),
					Args:        parts[7:13],
				},
			}
			isValid = true
		} else if eventType == "X" && len(parts) >= 7 {
			logEntry = TraceLog{
				EventType: "EXIT",
				FuncName:  funcName,
				Timestamp: time.Now().Format("2006-01-02 15:04:05.000000"),
				PID:       parts[2],
				TID:       parts[3],
				Data: ExitData{
					DurationNs:  parts[4],
					BackToRaw:   parts[5],
					BackToClean: cleanSymbol(parts[5]),
					ReturnValue: parts[6],
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
