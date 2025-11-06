package decoder

import (
	"fmt"
	"math"
	"sync"

	"monad-flow/model"
	"monad-flow/util"

	lru "github.com/hashicorp/golang-lru/v2"
)

const recentlyDecodedCacheSize = 1000
const maxRedundancyFactor = 7

type DecodedMessage struct {
	AppMessageHash [20]byte
	Data           []byte
}

type DecoderCache struct {
	pendingDecoders map[[20]byte]*managedDecoder
	recentlyDecoded *lru.Cache[[20]byte, bool]
	mu              sync.RWMutex
}

func NewDecoderCache() *DecoderCache {
	lruCache, err := lru.New[[20]byte, bool](recentlyDecodedCacheSize)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize LRU cache: %v", err))
	}

	return &DecoderCache{
		pendingDecoders: make(map[[20]byte]*managedDecoder),
		recentlyDecoded: lruCache, // [수정]
	}
}

func (dc *DecoderCache) HandleChunk(chunk *model.MonadChunkPacket) (*DecodedMessage, error) {
	// 1. 키(AppMessageHash)를 가져옵니다.
	key := chunk.AppMessageHash

	if dc.recentlyDecoded.Contains(key) {
		return nil, nil
	}

	// 2. 읽기 락(R-Lock)을 걸고 디코더가 이미 있는지 확인합니다.
	dc.mu.RLock()
	decoder, exists := dc.pendingDecoders[key]
	dc.mu.RUnlock()

	// 3. [시나리오 A] 디코더가 없으면 새로 생성합니다.
	if !exists {
		t := len(chunk.Payload)
		if t == 0 {
			return nil, fmt.Errorf("cannot decode chunk with zero-length payload (hash: %x)", key)
		}

		totalSize := chunk.AppMessageLen
		k := int(math.Ceil(float64(totalSize) / float64(t)))

		if k < util.SourceSymbolsMin {
			k = util.SourceSymbolsMin
		}

		encodedSymbolCapacity := k * maxRedundancyFactor
		if k > (util.SourceSymbolsMax / maxRedundancyFactor) {
			encodedSymbolCapacity = util.SourceSymbolsMax
		}

		// 쓰기 락(W-Lock)을 걸고 디코더를 생성/등록합니다.
		dc.mu.Lock()
		decoder, exists = dc.pendingDecoders[key]
		if !exists {
			newDecoder, err := newManagedDecoder(k, t, totalSize, encodedSymbolCapacity)
			if err != nil {
				dc.mu.Unlock()
				return nil, fmt.Errorf("failed to create new decoder for hash %x (K=%d, T=%d): %w", key, k, t, err)
			}
			decoder = newDecoder
			dc.pendingDecoders[key] = decoder
		}
		// 쓰기 락 해제
		dc.mu.Unlock()
	}

	// 4. [시나리오 B] 이제 'decoder' 변수(기존 것이든 새로 만든 것이든)가
	err := decoder.ReceiveSymbol(chunk.Payload, chunk.ChunkID)
	if err != nil {
		return nil, fmt.Errorf("failed to receive symbol %d for hash %x: %w", chunk.ChunkID, key, err)
	}

	// 5. 디코딩을 시도합니다.
	isDone, err := decoder.TryDecode()
	if err != nil {
		return nil, fmt.Errorf("decode attempt failed for hash %x: %w", key, err)
	}

	// 6. 디코딩이 완료되었는지 확인합니다.
	if isDone {
		data, err := decoder.ReconstructData()
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct data for hash %x: %w", key, err)
		}

		result := &DecodedMessage{
			AppMessageHash: key,
			Data:           data,
		}

		dc.mu.Lock()
		delete(dc.pendingDecoders, key)
		dc.mu.Unlock()

		dc.recentlyDecoded.Add(key, true)
		return result, nil
	}
	return nil, nil
}
