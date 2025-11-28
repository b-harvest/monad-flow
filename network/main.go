package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"monad-flow/tcp"
	"monad-flow/udp"
	"monad-flow/parser"
	"monad-flow/util"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/cilium/ebpf/ringbuf"
	"github.com/joho/godotenv"
	"github.com/zishang520/socket.io/clients/engine/v3/transports"
	"github.com/zishang520/socket.io/clients/socket/v3"
	"github.com/zishang520/socket.io/v3/pkg/types"
)

func main() {
	mtu := getMTU()

	client, err := connectSocketIO()
	if err != nil {
		log.Fatalf("Failed to connect to Socket.IO: %v", err)
	}

	if client != nil {
		defer (*client).Close()
	}

	var clientMutex sync.Mutex

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

	// Initialize Managers
	tcpManager := tcp.NewManager(ctx, &wg, client, &clientMutex)
	tcpManager.Start()

	udpManager := udp.NewManager(ctx, &wg, client, &clientMutex, mtu)
	udpManager.Start()

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
			captureTime := time.Now()
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
				tcpManager.HandlePacket(&packet)
			} else if packet.UDPLayer != nil {
				udpManager.HandlePacket(packet, int(realLen), captureTime)
			}
		}
	}()

	<-ctx.Done()

	log.Println("Stopping parser...")
	monitor.Close()

	tcpManager.Close()
	// udpManager doesn't have Close() but it listens to ctx.Done()

	log.Println("Waiting for workers...")
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

func connectSocketIO() (*socket.Socket, error) {
	godotenv.Load()
	sioURL := os.Getenv("BACKEND_URL")

	if sioURL == "" {
		log.Println("BACKEND_URL not set in .env, using default http://127.0.0.1:3000")
		sioURL = "http://127.0.0.1:3000"
	}

	connectChan := make(chan struct{})
	errChan := make(chan error, 1)

	var connectOnce sync.Once
	if sioURL == "no" {
		close(connectChan)
		return nil, nil
	}

	opts := socket.DefaultOptions()
	opts.SetTransports(types.NewSet(
		transports.WebSocket,
	))

	client, err := socket.Connect(sioURL, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate Socket.IO connection: %w", err)
	}

	client.On("connect", func(data ...any) {
		log.Printf("Socket.IO connected successfully! ID: %s", client.Id())
		connectOnce.Do(func() {
			close(connectChan)
		})
	})

	client.On("connect_error", func(err ...any) {
		log.Printf("[Socket.IO CONNECT_ERROR] %v", err[0])
		select {
		case errChan <- fmt.Errorf("socket.io connect_error: %v", err[0]):
		default:
		}
	})

	select {
	case <-connectChan:
		return client, nil

	case err := <-errChan:
		return nil, err

	case <-time.After(10 * time.Second):
		client.Close()
		return nil, errors.New("socket.io connection timed out")
	}
}
