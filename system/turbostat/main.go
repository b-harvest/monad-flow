package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"
)

type TurbostatMetric struct {
	Core    string  `json:"core"`     // "0", "1" 또는 "-" (Summary)
	CPU     string  `json:"cpu"`      // "0", "1" 또는 "-"
	AvgMHz  float64 `json:"avg_mhz"`  // Avg_MHz
	BusyPct float64 `json:"busy_pct"` // Busy%
	BzyMHz  float64 `json:"bzy_mhz"`  // Bzy_MHz
	TSCMHz  float64 `json:"tsc_mhz"`  // TSC_MHz
	IPC     float64 `json:"ipc"`      // IPC
	IRQ     int     `json:"irq"`      // IRQ
	CorWatt float64 `json:"cor_watt"` // CorWatt
	PkgWatt float64 `json:"pkg_watt"` // PkgWatt
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("사용법: sudo go run . <PID> (PID는 현재 코드에서 사용되지 않지만 인자로 받음)")
	}

	cmd := exec.Command("turbostat", "--interval", "0.5")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("StdoutPipe 연결 실패: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("StderrPipe 연결 실패: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("명령어 시작 실패: %v", err)
	}

	log.Println("... turbostat 파싱 시작 ...")

	go func() {
		io.Copy(os.Stderr, stderr)
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 10 {
			continue
		}

		if fields[0] == "Core" {
			continue
		}

		if !isDataRow(fields[0]) {
			continue
		}

		metric, err := parseTurbostatLine(fields)
		if err != nil {
			log.Printf("파싱 에러 (%s): %v", line, err)
			continue
		}

		jsonData, _ := json.Marshal(metric)
		fmt.Printf("[Parsed] %s\n", string(jsonData))
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("turbostat 종료: %v", err)
	}
}

func isDataRow(firstField string) bool {
	if firstField == "-" {
		return true
	}
	for _, char := range firstField {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func parseTurbostatLine(fields []string) (TurbostatMetric, error) {
	if len(fields) < 10 {
		return TurbostatMetric{}, fmt.Errorf("필드 개수 부족: %d", len(fields))
	}

	m := TurbostatMetric{
		Core: fields[0],
		CPU:  fields[1],
	}

	var err error

	parseFloat := func(s string) float64 {
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}

	parseInt := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
	}

	m.AvgMHz = parseFloat(fields[2])
	m.BusyPct = parseFloat(fields[3])
	m.BzyMHz = parseFloat(fields[4])
	m.TSCMHz = parseFloat(fields[5])
	m.IPC = parseFloat(fields[6])
	m.IRQ = parseInt(fields[7])
	m.CorWatt = parseFloat(fields[8])
	m.PkgWatt = parseFloat(fields[9])

	return m, err
}
