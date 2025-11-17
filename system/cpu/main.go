package main

import (
	"fmt"
	"io" // io.Copy를 위해 임포트
	"log"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("turbostat", "--interval", "0.5")

	// 1. (수정) turbostat의 "데이터"가 나오는 stdout 파이프를 연결합니다.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("StdoutPipe 연결 실패: %v", err)
	}

	// 2. (추가) 헤더 정보나 오류 메시지를 보기 위해 stderr도 연결합니다.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("StderrPipe 연결 실패: %v", err)
	}

	// 3. 명령어 시작
	if err := cmd.Start(); err != nil {
		log.Fatalf("turbostat 명령어 시작 실패: %v", err)
	}

	log.Println("... 'turbostat --interval 0.5'를 Go 프로그램으로 실행합니다 ...")
	log.Println("... (참고: 이 Go 프로그램은 반드시 'sudo'로 실행해야 합니다) ...")
	log.Println("... Ctrl+C로 Go 프로그램을 종료하면 turbostat도 함께 종료됩니다 ...")
	fmt.Println()

	// 4. (수정) bufio.Scanner 대신 goroutine과 io.Copy 사용
	//    stdout 데이터를 Go 프로그램의 stdout으로 "있는 그대로" 복사합니다.
	//    이 방식은 '\n'이든 '\r'이든 상관없이 동작합니다.
	go func() {
		io.Copy(os.Stdout, stdout)
	}()

	// 5. (추가) stderr 데이터도 Go 프로그램의 stderr로 "있는 그대로" 복사합니다.
	go func() {
		io.Copy(os.Stderr, stderr)
	}()

	// 6. 'turbostat' 프로세스가 종료될 때까지 대기
	if err := cmd.Wait(); err != nil {
		// Ctrl+C로 종료 시 "signal: interrupt" 오류가 발생할 수 있으며, 이는 정상입니다.
		log.Printf("turbostat 명령어가 종료되었습니다 (오류: %v)", err)
	}

	log.Println("... 모니터링 종료 ...")
}
