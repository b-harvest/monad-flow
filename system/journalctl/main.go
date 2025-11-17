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
	unitName := os.Args[1] // "monad-bft.service"

	j, err := sdjournal.NewJournal()
	if err != nil {
		log.Fatalf("저널을 여는데 실패했습니다: %v", err)
	}
	defer j.Close()

	// 1. _SYSTEMD_UNIT 필터 사용
	match := sdjournal.Match{
		Field: "_SYSTEMD_UNIT",
		Value: unitName,
	}
	if err = j.AddMatch(match.String()); err != nil {
		log.Fatalf("필터 추가에 실패했습니다 (%s): %v", match.String(), err)
	}

	// 2. [수정된 부분!]
	// 'SeekRealtime' (X) -> 'SeekRealtimeUsec' (O)

	// 30초 전 시간 계산
	thirtySecondsAgo := time.Now().Add(-30 * time.Second)

	// 'time.Time' 객체를 'uint64 마이크로초'로 변환
	// (time.UnixNano()는 나노초이므로 1000으로 나누어 마이크로초로 만듭니다)
	seekTimeUsec := uint64(thirtySecondsAgo.UnixNano() / 1000)

	// 올바른 메서드 이름과 인자 타입(uint64)으로 호출
	if err = j.SeekRealtimeUsec(seekTimeUsec); err != nil {
		log.Fatalf("30초 전 로그 시점으로 이동하는데 실패했습니다: %v", err)
	}

	fmt.Printf("==> [%s] 유닛의 최근 30초 로그 및 실시간 로그를 스트리밍합니다...\n", unitName)

	// 3. 실시간 스트리밍 루프 (이하 동일)
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
			ts := time.Unix(0, int64(entry.RealtimeTimestamp)*1000) // 마이크로초 -> 나노초
			fmt.Printf("%s: %s\n", ts.Format("Nov 02 15:04:05"), message)
		}
	}
}
