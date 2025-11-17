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
	"monad-flow/model"
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
	var udpMutex sync.Mutex

	tcpChan := make(chan *model.Packet, 10000)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("[TCP Worker] Started.")
		for {
			select {
			case <-ctx.Done():
				log.Println("[TCP Worker] Shutting down.")
				return
			case packet := <-tcpChan:
				assemblerMutex.Lock()
				assembler.AssembleWithTimestamp(
					packet.IPv4Layer.NetworkFlow(),
					packet.TCPLayer,
					time.Now(),
				)
				assemblerMutex.Unlock()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("[Ticker] Shutting down.")
				return
			case <-ticker.C:
				assemblerMutex.Lock()
				flushed, closed := assembler.FlushOlderThan(time.Now().Add(-1 * time.Second))
				assemblerMutex.Unlock()
				if flushed > 0 || closed > 0 {
					log.Printf("[TCP Reassembly] Flush: %d flushed, %d closed", flushed, closed)
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
				log.Printf("Error reading ringbuf: %v", err)
				continue
			}

			if ctx.Err() != nil {
				continue
			}

			if len(record.RawSample) < 4 {
				log.Println("Received invalid sample (too small)")
				continue
			}
			realLen := binary.LittleEndian.Uint32(record.RawSample[0:4])
			pktData := record.RawSample[4:]
			dumpLen := int(realLen)
			if dumpLen > len(pktData) {
				dumpLen = len(pktData)
			}

			packet := parser.ParsePacket(pktData[:dumpLen])

			if packet.TCPLayer != nil {
				select {
				case tcpChan <- &packet:
				default:
					log.Println("[WARN] TCP Channel full, dropping packet")
				}
			} else if packet.UDPLayer != nil {
				l7Payload := packet.Payload
				currentMTU := mtu
				currentRealLen := int(realLen)
				currentIPv4Len := int(packet.IPv4Layer.Length)
				go processUdpPacket(decoderCache, &udpMutex, l7Payload, currentMTU, currentRealLen, currentIPv4Len)
			}
		}
	}()

	<-ctx.Done()

	log.Println("Stopping parser...")
	monitor.Close()
	log.Println("Waiting for workers...")
	wg.Wait()
	log.Println("Closing TCP streams...")
	assemblerMutex.Lock()
	assembler.FlushOlderThan(time.Now())
	assemblerMutex.Unlock()
	log.Println("Shutdown complete.")
}

func processUdpPacket(
	decoderCache *decoder.DecoderCache,
	udpMutex *sync.Mutex,
	l7Payload []byte,
	mtu int,
	realLen int,
	ipv4Len int,
) {
	if l7Payload == nil {
		return
	}

	stride := mtu - (realLen - len(l7Payload) - (realLen - ipv4Len))
	if stride <= 0 {
		log.Printf("Invalid stride : %d", stride)
		return
	}

	offset := 0
	for offset < len(l7Payload) {
		remainingLen := len(l7Payload) - offset
		currentStride := stride

		if remainingLen < currentStride {
			currentStride = remainingLen
		}

		chunkData := l7Payload[offset : offset+currentStride]
		offset += currentStride

		udpMutex.Lock()
		data, err := processChunk(decoderCache, chunkData)
		if err != nil {
			log.Printf("Failed to process chunk: %v", err)
			udpMutex.Unlock()
			continue
		}
		if data != nil {
			if err := parser.HandleDecodedMessage(data); err != nil {
				log.Printf("[RLP-ERROR] Failed to decode message: %v", err)
			}
		}
		udpMutex.Unlock()
	}
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
