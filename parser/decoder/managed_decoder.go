package decoder

import (
	"fmt"
	"monad-flow/util"

	"github.com/bits-and-blooms/bitset"
)

// managedDecoder는 단일 AppMessage의 디코딩 상태를 관리합니다.
// Rust의 ManagedDecoder와 DecoderState의 역할을 일부 겸합니다.
type managedDecoder struct {
	params    *util.CodeParameters // K로부터 계산된 파라미터 (params.go)
	decoder   *lowLevelDecoder     // "두뇌" (행렬 연산)
	bufferSet *bufferSet           // "손" (바이트 버퍼 관리) (buffer.go)

	K         int    // 원본 심볼 수
	T         int    // 심볼 크기 (바이트)
	totalSize uint32 // 원본 데이터 총 크기 (app_message_len)

	// 수신된 청크 ID를 추적하여 중복을 방지합니다 (Rust의 seen_esis).
	seenSymbols    *bitset.BitSet
	symbolCapacity int // 최대 심볼 ID (K * 2로 가정)
}

// newManagedDecoder는 DecoderCache에 의해 호출됩니다 (첫 패킷 수신 시).
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
	// '두뇌'가 요청한 임시 버퍼 개수를 가져옵니다. (이것이 올바른 위치)
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
		seenSymbols: bitset.New(initialCapacity), // [수정]
		// [삭제] symbolCapacity 필드 제거
	}
	return md, nil
}

// [!!] ReceiveSymbol 함수 전체를 이 코드로 교체합니다.
// (Rust의 `received_encoded_symbol` 로직을 1:1로 포팅)
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

	// --- [!!] 버그 수정: Rust 로직과 동일하게 변경 ---

	// 4. "손"(bufferSet)에 새 버퍼를 *먼저* 추가합니다.
	// (이 단계에서 `bufferSet.buffers` 슬라이스가 실제로 확장됩니다.)
	_, err := md.bufferSet.addReceiveBuffer(payload)
	if err != nil {
		// (e.g., payload 크기가 T와 다를 경우)
		return err
	}

	// 5. '두뇌'가 요청할 XOR 콜백 정의
	//    이 콜백은 `lowLevelDecoder` 내부의 모든 함수(ReceiveSymbol, tryPeel 등)에서 사용됩니다.
	numTemp := md.decoder.numTempBuffersRequired()

	xorCallback := func(aFlat, bFlat int) error {
		// '두뇌'가 요청한 flat index를 '손'이 사용하는 bufferId로 변환
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
	// --- [!!] 수정 끝 ---

	// 6. "두뇌" (lowLevelDecoder)에 알림
	//    이제 "두뇌"가 `ReceiveSymbol` 내부에서 `xorBuffers`를 호출해도
	//    "손"(`bufferSet`)에는 모든 버퍼가 준비되어 있습니다.
	err = md.decoder.ReceiveSymbol(id, xorCallback)
	if err != nil {
		return fmt.Errorf("lowLevelDecoder.ReceiveSymbol: %w", err)
	}

	// 7. 수신 완료 처리
	md.seenSymbols.Set(idUint)
	return nil
}

// [!!] TryDecode 함수의 콜백도 `ReceiveSymbol`과 동일한 로직을 사용하도록 수정합니다.
func (md *managedDecoder) TryDecode() (bool, error) {
	// (inactivationThreshold 계산은 동일)
	const inactivationMultiplier = 384
	const inactivationShift = 8
	inactivationThreshold := (inactivationMultiplier * md.K) >> inactivationShift
	if inactivationThreshold < md.K {
		inactivationThreshold = md.K
	}

	// '두뇌'가 요청할 XOR 콜백 정의 (ReceiveSymbol의 콜백과 동일)
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

	// '두뇌'에게 디코딩 알고리즘 실행을 요청
	return md.decoder.TryDecode(inactivationThreshold, xorCallback)
}

// ReconstructData는 디코딩 완료 후 원본 데이터를 조립합니다.
// (Rust의 managed_decoder.reconstruct_source_data)
func (md *managedDecoder) ReconstructData() ([]byte, error) {
	// 1. "두뇌"에게 복원된 심볼 맵을 요청합니다.
	//    (Key: Source Symbol Index 0..K-1, Value: bufferId)
	symbolMap, err := md.decoder.GetReconstructedOrder()
	if err != nil {
		return nil, fmt.Errorf("GetReconstructedOrder failed: %w", err)
	}

	// 2. 최종 데이터를 조립할 버퍼 생성
	// (미리 전체 크기를 할당하여 append 오버헤드를 줄입니다)
	finalData := make([]byte, 0, md.totalSize)

	// 3. 0부터 K-1까지 순서대로 "손"에게 데이터 요청
	for i := 0; i < md.K; i++ {
		bufID, ok := symbolMap[i] // i = Source Symbol Index (0..K-1)
		if !ok {
			return nil, fmt.Errorf("%w: missing source symbol %d", ErrReconstruction, i)
		}

		// "손" (bufferSet)에서 실제 바이트 슬라이스 가져오기
		symbolData, err := md.bufferSet.buffer(bufID)
		if err != nil {
			return nil, fmt.Errorf("%w: bufferSet.buffer failed for symbol %d (bufID: %v): %w", ErrReconstruction, i, bufID, err)
		}

		// 4. 조립
		finalData = append(finalData, symbolData...)
	}

	// 5. 패딩 제거 (중요)
	// 원본 데이터(totalSize)가 심볼 크기(T)로 정확히 나누어 떨어지지 않으면,
	// 마지막 심볼에 패딩(쓰레기 값)이 포함됩니다.
	// `totalSize` (AppMessageLen)를 기준으로 정확히 잘라내야 합니다.
	if uint32(len(finalData)) < md.totalSize {
		return nil, fmt.Errorf("%w: final data length %d is less than expected %d",
			ErrReconstruction, len(finalData), md.totalSize)
	}

	return finalData[:md.totalSize], nil
}
