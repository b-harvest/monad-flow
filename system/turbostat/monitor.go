package turbostat

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type TurbostatMetric struct {
	Timestamp string  `json:"timestamp"`
	Core      string  `json:"core"`
	CPU       string  `json:"cpu"`
	AvgMHz    float64 `json:"avg_mhz"`
	BusyPct   float64 `json:"busy_pct"`
	BzyMHz    float64 `json:"bzy_mhz"`
	TSCMHz    float64 `json:"tsc_mhz"`
	IPC       float64 `json:"ipc"`
	IRQ       int     `json:"irq"`
	CorWatt   float64 `json:"cor_watt"`
	PkgWatt   float64 `json:"pkg_watt"`
}

func Start(ctx context.Context, wg *sync.WaitGroup, outChan chan<- TurbostatMetric) {
	defer wg.Done()

	path, err := exec.LookPath("turbostat")
	if err != nil {
		log.Println("[Turbostat] Command not found. (kernel-tools installed?)")
		return
	}

	cmd := exec.CommandContext(ctx, path, "--interval", "0.5")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[Turbostat] Stdout pipe error: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[Turbostat] Start error (Root required?): %v\n", err)
		return
	}

	fmt.Println("[Turbostat] Started Power Monitor")

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

		now := time.Now().Format("2006-01-02 15:04:05.000")
		metric, err := parseTurbostatLine(fields, now)
		if err != nil {
			continue
		}

		select {
		case outChan <- metric:
		case <-ctx.Done():
			return
		}
	}

	cmd.Wait()
	fmt.Println("[Turbostat] Service stopped.")
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

func parseTurbostatLine(fields []string, timestamp string) (TurbostatMetric, error) {
	if len(fields) < 10 {
		return TurbostatMetric{}, fmt.Errorf("field count mismatch")
	}

	m := TurbostatMetric{
		Timestamp: timestamp,
		Core:      fields[0],
		CPU:       fields[1],
	}

	// Helper closure for safe parsing
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

	return m, nil
}
