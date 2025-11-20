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
	uprobe:%s:%s {
		@start[tid] = nsecs;
		
		// E|PID|TID|Time|Target|Caller|Arg0|Arg1|Arg2|Arg3|Arg4|Arg5
		printf("E|%%d|%%d|%%s|%%s|%%s|0x%%lx|0x%%lx|0x%%lx|0x%%lx|0x%%lx|0x%%lx\n", 
			pid, tid, strftime("%%H:%%M:%%S.%%f", nsecs), 
			usym(reg("ip")), usym(*reg("sp")),
			arg0, arg1, arg2, arg3, arg4, arg5);
	}

	uretprobe:%s:%s {
		if (@start[tid]) {
			$dur_ns = nsecs - @start[tid];
			// X|PID|TID|DurationNS|BackToFunc|RetVal
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

	fmt.Printf("Tracing %s (Offset %s)...\n", TargetBinary, TargetOffset)
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

		if eventType == "E" && len(parts) >= 12 {
			pid := parts[1]
			tid := parts[2]
			timeStr := parts[3]
			targetRaw := parts[4]
			callerRaw := parts[5]

			targetClean := cleanSymbol(targetRaw)
			callerClean := cleanSymbol(callerRaw)

			args := parts[6:12]

			fmt.Printf("[ENTER] [%s] PID:%s TID:%s\n", timeStr, pid, tid)

			fmt.Printf("        Func (Raw):   %s\n", targetRaw)
			fmt.Printf("        Func (Clean): %s\n", targetClean)

			fmt.Printf("        Caller (Raw):   %s\n", callerRaw)
			fmt.Printf("        Caller (Clean): %s\n", callerClean)

			fmt.Println("        Arguments (Raw Hex):")
			for i, arg := range args {
				fmt.Printf("          Arg %d: %s\n", i+1, arg)
			}
			fmt.Println()

		} else if eventType == "X" && len(parts) >= 6 {
			pid := parts[1]
			tid := parts[2]
			durNs := parts[3]
			backToRaw := parts[4]
			retVal := parts[5]

			backToClean := cleanSymbol(backToRaw)

			fmt.Printf("[EXIT ] PID:%s TID:%s\n", pid, tid)
			fmt.Printf("        Took:   %s ns\n", durNs)
			fmt.Printf("        BackTo (Clean): %s\n", backToClean)
			fmt.Printf("        BackTo (Raw):   %s\n", backToRaw)
			fmt.Printf("        Return: %s\n", retVal)
			fmt.Println("----------------------------------------------------------------------------------")
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
