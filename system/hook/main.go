package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	TargetBinary = "/usr/local/bin/monad-node"
	TargetOffset = "0x96C720"
)

func main() {
	bpftraceScript := fmt.Sprintf(`
	/* 1. 함수 진입 (Enter) */
	uprobe:%s:%s {
		@start[tid] = nsecs;
		// 포맷: E|PID|TID|Time|TargetFunc|CallerFunc
		printf("E|%%d|%%d|%%s|%%s|%%s\n", 
			pid, tid, strftime("%%H:%%M:%%S.%%f", nsecs), usym(reg("ip")), usym(*reg("sp")));
	}

	/* 2. 함수 탈출 (Return) */
	uretprobe:%s:%s {
		if (@start[tid]) {
			$dur_ns = nsecs - @start[tid];
			// 포맷: X|PID|TID|DurationNS|BackToFunc
			printf("X|%%d|%%d|%%d|%%s\n", 
				pid, tid, $dur_ns, usym(reg("ip")));
			
			delete(@start[tid]);
		}
	}
	`, TargetBinary, TargetOffset, TargetBinary, TargetOffset)

	cmd := exec.Command("bpftrace", "-e", bpftraceScript)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe: %v\n", err)
		return
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting bpftrace: %v\n", err)
		fmt.Println("Hint: sudo 권한으로 실행했는지 확인해주세요.")
		return
	}

	fmt.Printf("Tracing Started...\nBinary: %s\nOffset: %s\n", TargetBinary, TargetOffset)
	fmt.Println("----------------------------------------------------------------------------------")

	scanner := bufio.NewScanner(stdout)
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

		if eventType == "E" && len(parts) >= 6 {
			pid := parts[1]
			tid := parts[2]
			timeStr := parts[3]
			targetFunc := parts[4]
			callerFunc := parts[5]

			fmt.Printf("[ENTER] [%s] PID:%s TID:%s\n", timeStr, pid, tid)
			fmt.Printf("        Target: %s\n", targetFunc)
			fmt.Printf("        Caller: %s\n", callerFunc)
			fmt.Println()

		} else if eventType == "X" && len(parts) >= 5 {
			pid := parts[1]
			tid := parts[2]
			durNs := parts[3]
			backToFunc := parts[4]

			fmt.Printf("[EXIT ] PID:%s TID:%s\n", pid, tid)
			fmt.Printf("        Took: %s ns\n", durNs)
			fmt.Printf("        BackTo: %s\n", backToFunc)
			fmt.Println("----------------------------------------------------------------------------------")
		}
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("bpftrace finished with error: %v\n", err)
	}
}
