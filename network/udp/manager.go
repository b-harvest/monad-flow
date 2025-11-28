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
				m.clientMutex.Lock()
				if m.client != nil {
					(*m.client).Emit(util.MONAD_CHUNK_EVENT, payload)
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

	jsonData, err := json.Marshal(chunk)
	if err != nil {
		log.Printf("JSON marshaling failed: %s", err)
		return nil, err
	}

	payload := map[string]interface{}{
		"type":      util.MONAD_CHUNK_PACKET_EVENT,
		"data":      json.RawMessage(jsonData),
		"timestamp": captureTime.UnixMicro(),
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
