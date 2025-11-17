package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("사용법: sudo go run . <PID>")
	}
	pid := os.Args[1]

	// 1. 'perf' 명령어 준비
	// -I 500 : 500ms (0.5초) 간격으로 통계를 새로고침하며 출력
	// -d     : 상세 카운터 (L1-dcache 등)
	// -p <PID>: 타겟 프로세스 ID
	cmd := exec.Command("perf", "stat", "-d", "-p", pid, "-I", "500")

	// 2. 'perf stat'은 출력을 stdout이 아닌 stderr(표준 오류)로 보냅니다.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	// 3. 명령어 시작
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	log.Printf("... PID %s 모니터링 시작 ('perf stat -I 500' 사용) ...\n", pid)
	log.Println("... Ctrl+C로 종료 ...")

	// 4. 스캐너를 사용해 stderr를 한 줄씩 실시간으로 읽어옵니다.
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		// perf의 출력을 그대로 Go 프로그램의 표준 출력으로 내보냅니다.
		fmt.Println(scanner.Text())
	}

	// 5. 'perf' 프로세스가 종료될 때까지 대기
	if err := cmd.Wait(); err != nil {
		log.Printf("perf 명령어가 오류와 함께 종료되었습니다: %v", err)
	}
}