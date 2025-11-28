package tcp

import (
	"context"
	"log"
	"sync"
	"time"

	"monad-flow/model"

	"github.com/google/gopacket/tcpassembly"
	"github.com/zishang520/socket.io/clients/socket/v3"
)

type Manager struct {
	ctx            context.Context
	wg             *sync.WaitGroup
	assembler      *tcpassembly.Assembler
	assemblerMutex sync.Mutex
	InputChan      chan *model.Packet
}

func NewManager(ctx context.Context, wg *sync.WaitGroup, client *socket.Socket, clientMutex *sync.Mutex) *Manager {
	streamFactory := &MonadTcpStreamFactory{Ctx: ctx, Client: client, ClientMutex: clientMutex}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	return &Manager{
		ctx:       ctx,
		wg:        wg,
		assembler: assembler,
		InputChan: make(chan *model.Packet, 10000),
	}
}

func (m *Manager) Start() {
	// TCP Worker
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		log.Println("[TCP Worker] Started.")
		for {
			select {
			case <-m.ctx.Done():
				log.Println("[TCP Worker] Shutting down.")
				return
			case packet := <-m.InputChan:
				m.assemblerMutex.Lock()
				m.assembler.AssembleWithTimestamp(
					packet.IPv4Layer.NetworkFlow(),
					packet.TCPLayer,
					time.Now(),
				)
				m.assemblerMutex.Unlock()
			}
		}
	}()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-m.ctx.Done():
				log.Println("[TCP Ticker] Shutting down.")
				return
			case <-ticker.C:
				m.assemblerMutex.Lock()
				flushed, closed := m.assembler.FlushOlderThan(time.Now().Add(-1 * time.Second))
				m.assemblerMutex.Unlock()
				if flushed > 0 || closed > 0 {
					log.Printf("[TCP Reassembly] Flush: %d flushed, %d closed", flushed, closed)
				}
			}
		}
	}()
}

func (m *Manager) Close() {
	log.Println("Closing TCP streams...")
	m.assemblerMutex.Lock()
	m.assembler.FlushOlderThan(time.Now())
	m.assemblerMutex.Unlock()
}

func (m *Manager) HandlePacket(packet *model.Packet) {
	select {
	case m.InputChan <- packet:
	default:
		log.Println("[WARN] TCP Channel full, dropping packet")
	}
}
