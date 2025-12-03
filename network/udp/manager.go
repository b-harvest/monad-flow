package udp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"monad-flow/decoder"
	"monad-flow/model"
	"monad-flow/parser"
	"monad-flow/util"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/zishang520/socket.io/clients/socket/v3"
)

type Manager struct {
	ctx          context.Context
	wg           *sync.WaitGroup
	decoderCache *decoder.DecoderCache
	client       *socket.Socket
	clientMutex  *sync.Mutex
	wsChan       chan map[string]interface{}
	mtu          int
	pingTargets  sync.Map
}

func NewManager(ctx context.Context, wg *sync.WaitGroup, client *socket.Socket, clientMutex *sync.Mutex, mtu int) *Manager {
	return &Manager{
		ctx:          ctx,
		wg:           wg,
		decoderCache: decoder.NewDecoderCache(),
		client:       client,
		clientMutex:  clientMutex,
		wsChan:       make(chan map[string]interface{}, 10000),
		mtu:          mtu,
	}
}

func (m *Manager) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		log.Println("[WS Worker] Started.")
		for {
			select {
			case <-m.ctx.Done():
				log.Println("[WS Worker] Shutting down.")
				return
			case payload := <-m.wsChan:
				eventName := util.MONAD_CHUNK_EVENT
				if msgType, ok := payload["type"].(int); ok {
					if msgType == util.PING_LATENCY_EVENT {
						eventName = util.PING_EVENT
					}
				}
				m.clientMutex.Lock()
				if m.client != nil {
					(*m.client).Emit(eventName, payload)
				}
				m.clientMutex.Unlock()
			}
		}
	}()
}

func (m *Manager) HandlePacket(packet model.Packet, realLen int, captureTime time.Time) {
	// Run in goroutine as per original main.go logic
	go func() {
		if packet.Payload == nil {
			return
		}

		ipv4Len := int(packet.IPv4Layer.Length)
		stride := m.mtu - (realLen - len(packet.Payload) - (realLen - ipv4Len))
		if stride <= 0 {
			log.Printf("Invalid stride : %d", stride)
			return
		}

		offset := 0
		for offset < len(packet.Payload) {
			remainingLen := len(packet.Payload) - offset
			currentStride := stride

			if remainingLen < currentStride {
				currentStride = remainingLen
			}

			chunkData := packet.Payload[offset : offset+currentStride]
			offset += currentStride

			payload, err := m.processChunk(packet, chunkData, captureTime)
			if err != nil {
				log.Printf("Failed to process chunk: %v", err)
				continue
			}

			if payload != nil {
				select {
				case m.wsChan <- payload:
				default:
					log.Println("[WARN] WS Channel full, dropping packet")
				}
			}
		}
	}()
}

func (m *Manager) processChunk(
	packet model.Packet,
	chunkData []byte,
	captureTime time.Time,
) (map[string]interface{}, error) {
	chunk, err := parser.ParseMonadChunkPacket(packet, chunkData)
	if err != nil {
		return nil, fmt.Errorf("chunk parsing failed: %w (data len: %d)", err, len(chunkData))
	}

	sourceIp := chunk.Network.Ipv4.SrcIp
	if sourceIp != "" {
		m.monitorLatency(sourceIp)
	}

	destinationIp := chunk.Network.Ipv4.DstIp
	if destinationIp != "" {
		m.monitorLatency(destinationIp)
	}

	senderInfo, err := util.RecoverSenderHybrid(chunk, chunkData)
	if err != nil {
		log.Printf("[Recovery-Warn] Failed to recover sender: %v", err)
	}

	jsonData, err := json.Marshal(chunk)
	if err != nil {
		log.Printf("JSON marshaling failed: %s", err)
		return nil, err
	}

	payload := map[string]interface{}{
		"type":        util.MONAD_CHUNK_PACKET_EVENT,
		"data":        json.RawMessage(jsonData),
		"timestamp":   captureTime.UnixMicro(),
		"secp_pubkey": senderInfo.NodeID,
	}

	decodedMsg, err := m.decoderCache.HandleChunk(chunk)
	if err != nil {
		if !errors.Is(err, decoder.ErrDuplicateSymbol) {
			return nil, fmt.Errorf("raptor processing error: %w", err)
		}
	}

	if decodedMsg != nil {
		appMessageHash := fmt.Sprintf("0x%x", decodedMsg.AppMessageHash)
		if err := parser.HandleDecodedMessage(decodedMsg.Data, appMessageHash); err != nil {
			log.Printf("[RLP-ERROR] Failed to decode message: %v", err)
		}
	}
	return payload, nil
}

func (m *Manager) monitorLatency(ip string) {
	if _, loaded := m.pingTargets.LoadOrStore(ip, true); loaded {
		return
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer m.pingTargets.Delete(ip)

		log.Printf("[Ping Monitor] Start monitoring: %s", ip)

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.sendPing(ip)
			}
		}
	}()
}

func (m *Manager) sendPing(ip string) {
	pinger, err := probing.NewPinger(ip)
	if err != nil {
		log.Printf("[Ping Error] Failed to init pinger for %s: %v", ip, err)
		return
	}
	pinger.SetPrivileged(true)
	pinger.Count = 1
	pinger.Timeout = 800 * time.Millisecond

	err = pinger.Run()
	if err != nil {
		log.Printf("[Ping Failed] IP: %s | Err: %v", ip, err)
		return
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv > 0 {
		m.wsChan <- map[string]interface{}{
			"type":      util.PING_LATENCY_EVENT,
			"ip":        ip,
			"rtt_ms":    float64(stats.AvgRtt.Microseconds()) / 1000.0,
			"timestamp": time.Now().Format("2006-01-02 15:04:05.000000"),
		}
	}
}
