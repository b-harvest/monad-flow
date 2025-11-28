package decoder

import (
	"fmt"
	"monad-flow/util"
	"sync"

	"github.com/bits-and-blooms/bitset"
)

type managedDecoder struct {
	params    *util.CodeParameters // K로부터 계산된 파라미터 (params.go)
	decoder   *lowLevelDecoder     // "두뇌" (행렬 연산)
	bufferSet *bufferSet           // "손" (바이트 버퍼 관리) (buffer.go)

	K         int    // 원본 심볼 수
	T         int    // 심볼 크기 (바이트)
	totalSize uint32 // 원본 데이터 총 크기 (app_message_len)

	seenSymbols    *bitset.BitSet
	symbolCapacity int // 최대 심볼 ID (K * 2로 가정)
	mu             sync.Mutex
}

func newManagedDecoder(k int, t int, totalSize uint32, capacity int) (*managedDecoder, error) {
	// 1. 파라미터 계산 (params.go)
	params, err := util.NewCodeParameters(k)
	if err != nil {
		return nil, fmt.Errorf("NewCodeParameters failed: %w", err)
	}

	// 2. "두뇌" 생성 (lowLevelDecoder)
	decoder, err := newLowLevelDecoder(params, capacity)
	if err != nil {
		return nil, fmt.Errorf("newLowLevelDecoder failed: %w", err)
	}

	// 3. "손" (BufferSet) 생성 (buffer.go)
	numTempBuffers := decoder.numTempBuffersRequired()
	bufferSet := newBufferSet(t, numTempBuffers)

	// 4. 수신 심볼 추적용 비트셋 초기화
	initialCapacity := uint(capacity)
	if initialCapacity < 32 {
		initialCapacity = 32
	}

	md := &managedDecoder{
		params:      params,
		decoder:     decoder,
		bufferSet:   bufferSet,
		K:           k,
		T:           t,
		totalSize:   totalSize,
		seenSymbols: bitset.New(initialCapacity),
	}
	return md, nil
}

func (md *managedDecoder) ReceiveSymbol(payload []byte, chunkID uint16) error {
	id := int(chunkID)
	idUint := uint(id)

	// 1. 유효성 검사 (ID 범위)
	if id >= 1_000_000 {
		return fmt.Errorf("%w: id %d seems unreasonably large", ErrInvalidSymbolID, id)
	}

	// 2. 유효성 검사 (중복)
	if md.seenSymbols.Test(idUint) {
		return fmt.Errorf("%w: id %d", ErrDuplicateSymbol, id)
	}

	// 3. 유효성 검사 (페이로드 크기)
	if len(payload) != md.T {
		return fmt.Errorf("%w: expected %d, got %d", ErrInvalidSymbolSize, md.T, len(payload))
	}

	// 4. "손"(bufferSet)에 새 버퍼를 *먼저* 추가합니다.
	// (이 단계에서 `bufferSet.buffers` 슬라이스가 실제로 확장됩니다.)
	_, err := md.bufferSet.addReceiveBuffer(payload)
	if err != nil {
		// (e.g., payload 크기가 T와 다를 경우)
		return err
	}

	// 5. '두뇌'가 요청할 XOR 콜백 정의
	// 이 콜백은 `lowLevelDecoder` 내부의 모든 함수(ReceiveSymbol, tryPeel 등)에서 사용됩니다.
	numTemp := md.decoder.numTempBuffersRequired()

	xorCallback := func(aFlat, bFlat int) error {
		var aID, bID bufferId

		if aFlat < numTemp {
			aID = bufferId{Type: tempBuffer, Index: aFlat}
		} else {
			aID = bufferId{Type: receiveBuffer, Index: aFlat - numTemp}
		}

		if bFlat < numTemp {
			bID = bufferId{Type: tempBuffer, Index: bFlat}
		} else {
			bID = bufferId{Type: receiveBuffer, Index: bFlat - numTemp}
		}

		return md.bufferSet.xorBuffers(aID, bID)
	}
	// 6. "두뇌" (lowLevelDecoder)에 알림
	err = md.decoder.ReceiveSymbol(id, xorCallback)
	if err != nil {
		return fmt.Errorf("lowLevelDecoder.ReceiveSymbol: %w", err)
	}

	// 7. 수신 완료 처리
	md.seenSymbols.Set(idUint)
	return nil
}

func (md *managedDecoder) TryDecode() (bool, error) {
	const inactivationMultiplier = 384
	const inactivationShift = 8
	inactivationThreshold := (inactivationMultiplier * md.K) >> inactivationShift
	if inactivationThreshold < md.K {
		inactivationThreshold = md.K
	}

	numTemp := md.decoder.numTempBuffersRequired()

	xorCallback := func(aFlat, bFlat int) error {
		var aID, bID bufferId

		if aFlat < numTemp {
			aID = bufferId{Type: tempBuffer, Index: aFlat}
		} else {
			aID = bufferId{Type: receiveBuffer, Index: aFlat - numTemp}
		}

		if bFlat < numTemp {
			bID = bufferId{Type: tempBuffer, Index: bFlat}
		} else {
			bID = bufferId{Type: receiveBuffer, Index: bFlat - numTemp}
		}

		return md.bufferSet.xorBuffers(aID, bID)
	}

	return md.decoder.TryDecode(inactivationThreshold, xorCallback)
}

func (md *managedDecoder) ReconstructData() ([]byte, error) {
	// 1. "두뇌"에게 복원된 심볼 맵을 요청합니다.
	//    (Key: Source Symbol Index 0..K-1, Value: bufferId)
	symbolMap, err := md.decoder.GetReconstructedOrder()
	if err != nil {
		return nil, fmt.Errorf("GetReconstructedOrder failed: %w", err)
	}

	// 2. 최종 데이터를 조립할 버퍼 생성
	finalData := make([]byte, 0, md.totalSize)

	// 3. 0부터 K-1까지 순서대로 "손"에게 데이터 요청
	for i := 0; i < md.K; i++ {
		bufID, ok := symbolMap[i] // i = Source Symbol Index (0..K-1)
		if !ok {
			return nil, fmt.Errorf("%w: missing source symbol %d", ErrReconstruction, i)
		}

		symbolData, err := md.bufferSet.buffer(bufID)
		if err != nil {
			return nil, fmt.Errorf("%w: bufferSet.buffer failed for symbol %d (bufID: %v): %w", ErrReconstruction, i, bufID, err)
		}

		// 4. 조립
		finalData = append(finalData, symbolData...)
	}

	// 5. 패딩 제거 (중요)
	if uint32(len(finalData)) < md.totalSize {
		return nil, fmt.Errorf("%w: final data length %d is less than expected %d",
			ErrReconstruction, len(finalData), md.totalSize)
	}

	return finalData[:md.totalSize], nil
}
