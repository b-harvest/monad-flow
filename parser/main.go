package main

import (
	"encoding/binary"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"monad-flow/parser"
	"monad-flow/util"

	"github.com/cilium/ebpf/ringbuf"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: sudo %s <interface-name>", os.Args[0])
	}
	ifName := os.Args[1]

	monitor, err := util.NewBPFMonitor(ifName)
	if err != nil {
		log.Fatalf("Failed to initialize eBPF monitor: %v", err)
	}
	defer monitor.Close()

	rd := monitor.RingBufReader

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Waiting for packets on port 8000...")
	log.Println("Run 'sudo cat /sys/kernel/debug/tracing/trace_pipe' to see kernel prints.")

	go func() {
		for {
			record, err := rd.Read()
			if err != nil {
				if errors.Is(err, ringbuf.ErrClosed) {
					log.Println("Ring buffer closed")
					return
				}
				log.Printf("Error reading from ring buffer: %v", err)
				continue
			}

			if len(record.RawSample) < 4 {
				log.Println("Received invalid sample (too small)")
				continue
			}

			// 1. 실제 패킷 길이(len 필드)를 읽습니다.
			realLen := binary.LittleEndian.Uint32(record.RawSample[0:4])

			// 2. 4바이트 이후의 데이터(data 필드)를 가져옵니다.
			pktData := record.RawSample[4:]

			// 3. 실제 덤프할 길이를 계산합니다.
			dumpLen := int(realLen)
			if dumpLen > len(pktData) {
				dumpLen = len(pktData)
			}
			packet := parser.ParsePacket(pktData[:dumpLen])
			chunk, err := parser.ParseMonadChunkPacket(packet.Payload)
			if err != nil {
				log.Printf("Chunk parsing failed: %v (data len: %d)", err, len(packet.Payload))
				return
			}
			util.PrintMonadPacketDetails(chunk)
		}
	}()

	<-stop
}
