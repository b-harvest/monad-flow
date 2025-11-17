package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runCommand(name string, args ...string) {
	fmt.Printf("\n--- [실행]: %s %v ---\n", name, args)

	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()

	if args[1] == "latency" || args[1] == "timehist" || name == "offcputime" {
		fmt.Println(string(output))
	}

	if err != nil {
		fmt.Printf("--- [에러]: %s 실행 실패: %v ---\n", name, err)
	}
}

func main() {
	// 1. 이 프로그램은 sudo/root 권한이 필요합니다.
	if os.Geteuid() != 0 {
		fmt.Println("이 프로그램은 sudo 또는 root 권한으로 실행해야 합니다.")
		fmt.Println("예: sudo go run .")
		os.Exit(1)
	}
	const perfDataFile = "./perf.data"

	// 2. [명령어 1] offcputime 실행
	runCommand("offcputime", "-p", "2052396", "1")

	// 3. [명령어 2] perf sched record 실행
	runCommand("perf", "sched", "record", "-p", "2052396", "-o", perfDataFile, "sleep", "0.1")

	// 4. [명령어 3] perf sched latency 실행
	runCommand("perf", "sched", "latency", "-i", perfDataFile)

	// 5. [명령어 4] perf sched timehist 실행
	runCommand("perf", "sched", "timehist", "-i", perfDataFile)

	fmt.Println("\n--- [모든 명령어 실행 완료] ---")
}
