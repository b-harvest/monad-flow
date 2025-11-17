package main

import (
	"fmt"
	"os"
	"os/exec"
	// Sleep을 사용하진 않지만, 혹시 나중에 짧게 쓸 수 있으니 놔둡니다.
)

func runCommand(name string, args ...string) {
	// ... (이전과 동일) ...
	fmt.Printf("\n--- [실행]: %s %v ---\n", name, args)

	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()

	if len(args) > 1 && (args[1] == "latency" || args[1] == "timehist" || name == "offcputime") {
		fmt.Println(string(output))
	}

	if err != nil {
		fmt.Printf("--- [에러]: %s 실행 실패: %v ---\n", name, err)
	}
}

func main() {
	// ... (권한 및 PID 인자 확인 부분은 이전과 동일) ...
	if os.Geteuid() != 0 {
		fmt.Println("이 프로그램은 sudo 또는 root 권한으로 실행해야 합니다.")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("모니터링할 PID를 인자로 전달해야 합니다.")
		fmt.Println("예: sudo go run . 12345")
		os.Exit(1)
	}
	pid := os.Args[1]
	fmt.Printf("--- [PID %s] 모니터링 시작 ---\n", pid)

	// 3. 무한 루프 시작
	for {
		fmt.Println("--- [새 사이클 시작] ---")

		// 4. [명령어 1] offcputime 실행 (1초 소요)
		runCommand("offcputime", "-p", pid, "1")

		fmt.Printf("\n--- [사이클 완료] 즉시 다음 사이클을 시작합니다 ... (Ctrl+C로 종료) ---\n\n")

		// 8. 2초 대기 라인 삭제!
		// time.Sleep(2 * time.Second)
	}
}
