package decoder

import (
	"fmt"
	"math"
	"sync"

	// 파싱된 청크 구조체를 사용하기 위해
	"monad-flow/model"
	"monad-flow/util"

	lru "github.com/hashicorp/golang-lru/v2"
)

// [수정] Rust의 `RECENTLY_DECODED_CACHE_SIZE`와 유사한 상수 정의
const recentlyDecodedCacheSize = 1000
const maxRedundancyFactor = 7

// DecodedMessage는 디코딩이 성공했을 때 반환되는 결과물입니다.
type DecodedMessage struct {
	AppMessageHash [20]byte
	Data           []byte
}

// DecoderCache는 모든 진행 중인 디코딩 작업을 관리합니다.
type DecoderCache struct {
	pendingDecoders map[[20]byte]*managedDecoder

	// [수정] 이미 디코딩된 메시지를 추적하여 중복 패킷을 빠르게 무시합니다.
	// Key: [20]byte (AppMessageHash), Value: bool (존재 여부만 중요)
	recentlyDecoded *lru.Cache[[20]byte, bool]

	mu sync.RWMutex
}

// NewDecoderCache는 새 디코더 캐시를 초기화합니다.
func NewDecoderCache() *DecoderCache {
	// [수정] LRU 캐시 초기화 로직 추가
	// New()는 에러를 반환할 수 있으나, 고정 사이즈(>0)이므로 패닉 처리
	lruCache, err := lru.New[[20]byte, bool](recentlyDecodedCacheSize)
	if err != nil {
		// 이 경우는 `size <= 0`일 때만 발생하므로, 상수인 경우 패닉이 타당합니다.
		panic(fmt.Sprintf("failed to initialize LRU cache: %v", err))
	}

	return &DecoderCache{
		pendingDecoders: make(map[[20]byte]*managedDecoder),
		recentlyDecoded: lruCache, // [수정]
	}
}

// HandleChunk는 main.go에서 호출할 단일 진입점입니다.
// 이 함수는 수신된 청크를 적절한 디코더에 전달하고,
// 디코딩이 완료되면 원본 메시지를 반환합니다.
func (dc *DecoderCache) HandleChunk(chunk *model.MonadChunkPacket) (*DecodedMessage, error) {
	// 1. 키(AppMessageHash)를 가져옵니다.
	key := chunk.AppMessageHash

	// --- [수정] 1. Recently Decoded 캐시 확인 ---
	// `pendingDecoders` 맵을 잠그기 전에 LRU 캐시를 먼저 확인합니다.
	// (lru.Cache는 내부적으로 스레드 안전합니다.)
	if dc.recentlyDecoded.Contains(key) {
		// 이미 디코딩이 완료된 메시지입니다.
		// 에러 없이 조용히 무시합니다.
		return nil, nil
	}
	// --- [수정] 끝 ---

	// 2. 읽기 락(R-Lock)을 걸고 디코더가 이미 있는지 확인합니다.
	dc.mu.RLock()
	decoder, exists := dc.pendingDecoders[key]
	dc.mu.RUnlock()

	// 3. [시나리오 A] 디코더가 없으면 새로 생성합니다.
	if !exists {
		// 이 청크가 이 메시지의 "첫 번째" 청크입니다.
		// K와 T를 이 청크의 메타데이터로부터 계산합니다.

		// T (Symbol Size) = 청크 페이로드의 길이
		t := len(chunk.Payload)
		if t == 0 {
			return nil, fmt.Errorf("cannot decode chunk with zero-length payload (hash: %x)", key)
		}

		// K (Num Source Symbols) = ceil(TotalSize / T)
		totalSize := chunk.AppMessageLen // uint32
		k := int(math.Ceil(float64(totalSize) / float64(t)))

		// K는 최소값(1) 이상이어야 합니다.
		if k < util.SourceSymbolsMin {
			k = util.SourceSymbolsMin
		}

		encodedSymbolCapacity := k * maxRedundancyFactor
		if k > (util.SourceSymbolsMax / maxRedundancyFactor) { // 오버플로우 방지
			encodedSymbolCapacity = util.SourceSymbolsMax
		}

		// 쓰기 락(W-Lock)을 걸고 디코더를 생성/등록합니다.
		dc.mu.Lock()
		// Double-check: RUnlock과 Lock 사이에 다른 고루틴이
		// 디코더를 이미 생성했을 수 있으므로 다시 확인합니다.
		decoder, exists = dc.pendingDecoders[key]
		if !exists {
			// [호출부 수정] capacity 인자 추가
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
	// 이 청크를 처리할 디코더를 가리킵니다.

	// (아직 우리가 구현하지 않은) ReceiveSymbol을 호출합니다.
	// 이 메서드는 내부적으로 중복 청크 검사 등을 수행해야 합니다.
	err := decoder.ReceiveSymbol(chunk.Payload, chunk.ChunkID)
	if err != nil {
		// 예: 중복 심볼(패킷), 잘못된 페이로드 크기 등
		// 이미 처리한 중복 패킷은 에러가 아니라 'nil, nil'을 반환할 수 있습니다.
		// 지금은 간단히 에러로 처리합니다.
		return nil, fmt.Errorf("failed to receive symbol %d for hash %x: %w", chunk.ChunkID, key, err)
	}

	// 5. 디코딩을 시도합니다.
	// (아직 우리가 구현하지 않은) TryDecode를 호출합니다.
	isDone, err := decoder.TryDecode()
	if err != nil {
		return nil, fmt.Errorf("decode attempt failed for hash %x: %w", key, err)
	}

	// 6. 디코딩이 완료되었는지 확인합니다.
	if isDone {
		// (아직 우리가 구현하지 않은) ReconstructData를 호출합니다.
		data, err := decoder.ReconstructData()
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct data for hash %x: %w", key, err)
		}

		// 최종 결과물을 생성합니다.
		result := &DecodedMessage{
			AppMessageHash: key,
			Data:           data,
		}

		// --- [수정] 맵에서 삭제하고, LRU 캐시에 추가 ---
		dc.mu.Lock()
		delete(dc.pendingDecoders, key)
		dc.mu.Unlock()

		dc.recentlyDecoded.Add(key, true)
		// --- [수정] 끝 ---

		return result, nil
	}

	// 아직 디코딩이 완료되지 않음 (더 많은 청크 필요)
	return nil, nil
}
