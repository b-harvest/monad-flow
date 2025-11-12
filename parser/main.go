package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"monad-flow/decoder"
	"monad-flow/parser"
	"monad-flow/util"

	"github.com/cilium/ebpf/ringbuf"
	"github.com/google/gopacket/tcpassembly"
	"github.com/joho/godotenv"
)

func main() {
	mtu := getMTU()

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

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("Ctrl-C received. Starting shutdown process...")
		cancel()
	}()

	decoderCache := decoder.NewDecoderCache()
	streamFactory := &decoder.MonadTcpStreamFactory{Ctx: ctx}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)
	var assemblerMutex sync.Mutex

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("[Ticker] Ticker goroutine shutting down.")
				return
			case <-ticker.C:
				assemblerMutex.Lock()
				flushed, closed := assembler.FlushOlderThan(time.Now().Add(-50 * time.Millisecond))
				assemblerMutex.Unlock()
				if flushed > 0 || closed > 0 {
					log.Printf("[TCP Reassembly] Assembler Flush: %d flushed, %d closed", flushed, closed)
				}
			}
		}
	}()

	log.Println("Waiting for packets on port 8000...")
	log.Println("Run 'sudo cat /sys/kernel/debug/tracing/trace_pipe' to see kernel prints.")

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Println("[eBPF] Context cancellation signal detected. Shutting down eBPF goroutine...")
				return
			default:
			}
			record, err := rd.Read()
			if err != nil {
				if errors.Is(err, ringbuf.ErrClosed) {
					log.Println("Ring buffer closed")
					return
				}
				select {
				case <-ctx.Done():
					return
				default:
				}
				log.Printf("Error reading from ring buffer: %v", err)
				continue
			}

			if ctx.Err() != nil {
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

			if packet.TCPLayer != nil {
				assemblerMutex.Lock()
				assembler.AssembleWithTimestamp(
					packet.IPv4Layer.NetworkFlow(),
					packet.TCPLayer,
					time.Now(),
				)
				assemblerMutex.Unlock()
			} else if packet.UDPLayer != nil {
				stride := mtu - (int(realLen) - len(packet.Payload) - (int(realLen) - int(packet.IPv4Layer.Length)))
				if stride <= 0 {
					log.Printf("Invalid stride : %d = %d - (%d - %d - (%d - %d))", stride, mtu, int(realLen), len(packet.Payload), int(realLen), int(packet.IPv4Layer.Length))
					log.Fatalf("Invalid MTU (%d), calculated stride is %d", mtu, stride)
				}

				if packet.Payload == nil {
					continue
				}

				l7Payload := packet.Payload
				offset := 0
				for offset < len(l7Payload) {
					remainingLen := len(l7Payload) - offset
					currentStride := stride // .env에서 계산한 Stride

					if remainingLen < currentStride {
						currentStride = remainingLen
					}

					chunkData := l7Payload[offset : offset+currentStride]
					offset += currentStride

					data, err := processChunk(decoderCache, chunkData)

					if err != nil {
						log.Printf("Failed to process chunk: %v", err)
						continue
					}
					if data != nil {
						if err := parser.HandleDecodedMessage(data); err != nil {
							log.Printf("[RLP-ERROR] Failed to decode message: %v", err)
						}
					}
				}
			}
		}
	}()

	<-ctx.Done()

	log.Println("Stopping Monad packet parser...")
	log.Println("Closing eBPF monitor...")
	if err := monitor.Close(); err != nil {
		log.Printf("Warning: error closing monitor: %v", err)
	}
	log.Println("Closing all active TCP streams forcefully...")
	assemblerMutex.Lock()
	flushed, closed := assembler.FlushOlderThan(time.Now())
	assemblerMutex.Unlock()
	log.Printf("[TCP Reassembly] Final Flush: %d flushed, %d closed", flushed, closed)
	log.Println("Waiting for all goroutines to stop...")
	wg.Wait()
	log.Println("Shutdown complete.")
}

func getMTU() int {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default MTU 1480")
	}

	mtuStr := os.Getenv("MTU")
	if mtuStr == "" {
		mtuStr = "1480"
	}

	mtu, err := strconv.Atoi(mtuStr)
	if err != nil {
		log.Fatalf("Invalid MTU value in .env: %s", mtuStr)
	}
	return mtu
}

func processChunk(decoderCache *decoder.DecoderCache, chunkData []byte) ([]byte, error) {
	chunk, err := parser.ParseMonadChunkPacket(chunkData)
	if err != nil {
		return nil, fmt.Errorf("chunk parsing failed: %w (data len: %d)", err, len(chunkData))
	}

	decodedMsg, err := decoderCache.HandleChunk(chunk)
	if err != nil {
		if !errors.Is(err, decoder.ErrDuplicateSymbol) {
			return nil, fmt.Errorf("raptor processing error: %w", err)
		}
	}

	if decodedMsg != nil {
		return decodedMsg.Data, nil
	}
	return nil, nil
}
