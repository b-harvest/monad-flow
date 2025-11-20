package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("사용법: sudo go run . <PID>")
	}
	pid := os.Args[1]

	cmd := exec.Command("perf", "stat", "-d", "-p", pid, "-I", "500")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	log.Printf("... PID %s 모니터링 시작 ...\n", pid)
	log.Println("... Ctrl+C로 종료 ...")

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	scanner := bufio.NewScanner(stderr)

	var lastTimestamp string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parsed, err := parsePerfLine(line)
		if err != nil {
			fmt.Println(line)
			continue
		}

		if parsed.Timestamp != lastTimestamp {
			w.Flush()
			fmt.Printf("\n[ Time: %s sec ]\n", parsed.Timestamp)
			fmt.Fprintln(w, "EVENT\tVALUE\tUNIT\tMETRIC")
			fmt.Fprintln(w, "-----\t-----\t----\t------")
			lastTimestamp = parsed.Timestamp
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", parsed.Event, parsed.Value, parsed.Unit, parsed.Metric)
	}

	w.Flush() // 루프 종료 후 잔여 버퍼 출력

	if err := cmd.Wait(); err != nil {
		log.Printf("perf 종료: %v", err)
	}
}

type PerfData struct {
	Timestamp string
	Value     string
	Unit      string
	Event     string
	Metric    string
}

func parsePerfLine(line string) (PerfData, error) {
	// 1. # 문자를 기준으로 데이터 부분(앞)과 메트릭 설명(뒤)을 분리
	parts := strings.SplitN(line, "#", 2)
	dataPart := strings.TrimSpace(parts[0])
	metricPart := ""
	if len(parts) > 1 {
		metricPart = strings.TrimSpace(parts[1])
	}

	// 2. 데이터 부분을 공백 기준으로 분리
	fields := strings.Fields(dataPart)
	if len(fields) < 1 {
		return PerfData{}, fmt.Errorf("invalid line")
	}

	// fields[0] 은 항상 Timestamp
	p := PerfData{
		Timestamp: fields[0],
		Metric:    metricPart,
	}

	// 데이터 필드가 부족한 경우 (예: 타임스탬프만 있고 값이 없는 경우)
	if len(fields) == 1 {
		p.Event = "(info)"
		return p, nil
	}

	// 3. 값과 이벤트명 파싱
	if strings.Contains(dataPart, "<not supported>") {
		p.Value = "Not Supported"
		p.Event = fields[len(fields)-1]
	} else {
		p.Value = fields[1]
		if len(fields) > 2 {
			if fields[2] == "msec" {
				p.Unit = "msec"
				if len(fields) > 3 {
					p.Event = fields[3]
				}
			} else {
				p.Event = fields[2]
			}
		}
	}

	return p, nil
}
