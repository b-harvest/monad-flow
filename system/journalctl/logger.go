package journalctl

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
)

type Config struct {
	Services []string // 예: ["monad-node.service", "monad-bft.service"]
}

type LogEntry struct {
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	PID       string `json:"pid,omitempty"`
}

func Start(ctx context.Context, wg *sync.WaitGroup, outChan chan<- LogEntry, cfg Config) {
	defer wg.Done()

	if len(cfg.Services) == 0 {
		log.Println("[Journalctl] No services configured to monitor.")
		return
	}

	for _, serviceName := range cfg.Services {
		wg.Add(1)
		go tailService(ctx, wg, outChan, serviceName)
	}
}

func tailService(ctx context.Context, wg *sync.WaitGroup, outChan chan<- LogEntry, unitName string) {
	defer wg.Done()

	j, err := sdjournal.NewJournal()
	if err != nil {
		log.Printf("[Journalctl] Failed to open journal for %s: %v\n", unitName, err)
		return
	}
	defer j.Close()

	match := sdjournal.Match{
		Field: "_SYSTEMD_UNIT",
		Value: unitName,
	}
	if err = j.AddMatch(match.String()); err != nil {
		log.Printf("[Journalctl] Failed to add match for %s: %v\n", unitName, err)
		return
	}

	if err = j.SeekTail(); err != nil {
		log.Printf("[Journalctl] Failed to seek to tail for %s: %v\n", unitName, err)
	}

	_, _ = j.Next()

	fmt.Printf("[Journalctl] Started monitoring (Real-time only): %s\n", unitName)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[Journalctl] Stopping monitor for: %s\n", unitName)
			return
		default:
			ret, err := j.Next()
			if err != nil {
				log.Printf("[Journalctl] Read error on %s: %v\n", unitName, err)
				time.Sleep(2 * time.Second)
				continue
			}

			if ret == 0 {
				j.Wait(1 * time.Second)
				continue
			}

			entry, err := j.GetEntry()
			if err != nil {
				continue
			}

			if message, ok := entry.Fields["MESSAGE"]; ok {
				ts := time.Unix(0, int64(entry.RealtimeTimestamp)*1000)
				pid := entry.Fields["_PID"]

				logData := LogEntry{
					Service:   unitName,
					Timestamp: ts.Format("15:04:05.000000"),
					Message:   message,
					PID:       pid,
				}

				select {
				case outChan <- logData:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
