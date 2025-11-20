// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("에러: systemd 유닛 이름을 인자로 제공해야 합니다. (예: monad-bft.service)")
		return
	}
	unitName := os.Args[1]

	j, err := sdjournal.NewJournal()
	if err != nil {
		log.Fatalf("저널을 여는데 실패했습니다: %v", err)
	}
	defer j.Close()

	match := sdjournal.Match{
		Field: "_SYSTEMD_UNIT",
		Value: unitName,
	}
	if err = j.AddMatch(match.String()); err != nil {
		log.Fatalf("필터 추가에 실패했습니다 (%s): %v", match.String(), err)
	}
	thirtySecondsAgo := time.Now().Add(-30 * time.Second)
	seekTimeUsec := uint64(thirtySecondsAgo.UnixNano() / 1000)

	if err = j.SeekRealtimeUsec(seekTimeUsec); err != nil {
		log.Fatalf("30초 전 로그 시점으로 이동하는데 실패했습니다: %v", err)
	}

	fmt.Printf("==> [%s] 유닛의 최근 30초 로그 및 실시간 로그를 스트리밍합니다...\n", unitName)

	for {
		r, err := j.Next()
		if err != nil {
			log.Fatalf("로그 항목을 가져오는데 실패했습니다: %v", err)
		}

		if r == 0 {
			j.Wait(time.Second)
			continue
		}

		entry, err := j.GetEntry()
		if err != nil {
			log.Fatalf("로그 엔트리를 읽는데 실패했습니다: %v", err)
		}

		if message, ok := entry.Fields["MESSAGE"]; ok {
			ts := time.Unix(0, int64(entry.RealtimeTimestamp)*1000)
			fmt.Printf("%s: %s\n", ts.Format("Nov 02 15:04:05"), message)
		}
	}
}
