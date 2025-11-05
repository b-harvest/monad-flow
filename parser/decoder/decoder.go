package decoder

import (
	"errors"
	"fmt"
	"monad-flow/util"
	"sort"
	// 비트맵 라이브러리 (중복 청크 확인용)
)

// --- 에러 정의 (Error Definitions) ---
var (
	ErrDuplicateSymbol   = errors.New("duplicate symbol")
	ErrInvalidSymbolSize = errors.New("invalid symbol size")
	ErrInvalidSymbolID   = errors.New("invalid symbol ID")
	ErrReconstruction    = errors.New("reconstruction failed")
	ErrDecodeNotDone     = errors.New("decoding is not yet complete")
)

// ========================================================================
//
//	lowLevelDecoder (Stub)
//	  "두뇌" - 실제 디코딩 알고리즘 (Peeling, Gaussian)
//
// ========================================================================

// [구조체 수정] Rust의 `Decoder`와 동일하게 필드 정의
type lowLevelDecoder struct {
	params *util.CodeParameters

	// `decoder_state.go`에서 포팅한 자료구조들
	bufferState             []*buffer             // Vec<Buffer>
	intermediateSymbolState []*intermediateSymbol // Vec<IntermediateSymbol>
	buffersActiveUsable     *bufferWeightMap      // BufferWeightMap (Min-Heap)
	buffersInactivated      *bufferWeightMap      // BufferWeightMap (Min-Heap)

	numRedundantBuffers    uint16
	numSourceSymbolsPaired int // K에 도달하면 디코딩 완료
}

// [함수 수정] newLowLevelDecoder 스텁을 실제 구현으로 교체
// (Rust의 `Decoder::with_capacity` 포팅)
func newLowLevelDecoder(params *util.CodeParameters, capacity int) (*lowLevelDecoder, error) {

	// --- Rust: `buffer_state`와 `intermediate_symbol_state` 초기화 ---
	// 제약 행렬(Constraint Matrix)의 행(S+H)과 열(K+S+H)을 생성합니다.

	numLdpcSymbols := int(params.NumLdpcSymbols)
	numHalfSymbols := int(params.NumHalfSymbols)
	numSourceSymbols := int(params.NumSourceSymbols)
	numIntermediateSymbols := int(params.NumIntermediateSymbols) // K+S+H

	// 1. buffer_state: S+H 개의 "임시 버퍼" 상태를 생성합니다.
	numTempBuffers := numLdpcSymbols + numHalfSymbols
	bufferState := make([]*buffer, numTempBuffers)
	for i := 0; i < numTempBuffers; i++ {
		bufferState[i] = newBuffer()
	}

	// 2. intermediate_symbol_state: K+S+H 개의 "중간 심볼" 상태를 생성합니다.
	intermediateSymbolState := make([]*intermediateSymbol, numIntermediateSymbols)
	for i := 0; i < numIntermediateSymbols; i++ {
		intermediateSymbolState[i] = newIntermediateSymbol()
	}

	// --- Rust: 행렬 채우기 (G_LDPC, I_S, G_Half, I_H) ---
	// 이 로직은 버퍼와 심볼 간의 초기 XOR 관계를 설정합니다.

	// 3. G_LDPC
	// ------------------------------------------------------------------
	params.GLdpc(func(bufferIndex, symbolIndex int) {
		bufIdx := uint16(bufferIndex)
		symIdx := uint16(symbolIndex)

		bufferState[bufIdx].appendIntermediateSymbolID(symIdx, true) // active_used_weight++
		_ = intermediateSymbolState[symIdx].activePush(bufIdx)
	})
	// ------------------------------------------------------------------

	// 4. I_S (Identity Matrix for S)
	for i := 0; i < numLdpcSymbols; i++ {
		bufIdx := uint16(i)
		symIdx := uint16(numSourceSymbols + i) // K + i

		bufferState[bufIdx].appendIntermediateSymbolID(symIdx, true) // active_used_weight++
		if err := intermediateSymbolState[symIdx].activePush(bufIdx); err != nil {
			// (이론상 newIntermediateSymbol이 Active로 생성하므로 에러 발생 안 함)
			return nil, fmt.Errorf("I_S activePush failed: %w", err)
		}
	}

	// 5. G_Half
	// ------------------------------------------------------------------
	params.GHalf(func(h, j int) { // h = 0..H-1, j = 0..K+S-1
		bufIdx := uint16(numLdpcSymbols + h) // S + h
		symIdx := uint16(j)

		bufferState[bufIdx].appendIntermediateSymbolID(symIdx, true) // active_used_weight++
		// (에러 처리는 activePush의 시그니처가 반환하지 않으므로 제거, 필요시 패닉)
		_ = intermediateSymbolState[symIdx].activePush(bufIdx)
	})

	// 6. I_H (Identity Matrix for H)
	for i := 0; i < numHalfSymbols; i++ {
		bufIdx := uint16(numLdpcSymbols + i)                    // S + i
		symIdx := uint16(numSourceSymbols + numLdpcSymbols + i) // K + S + i

		bufferState[bufIdx].appendIntermediateSymbolID(symIdx, true) // active_used_weight++
		if err := intermediateSymbolState[symIdx].activePush(bufIdx); err != nil {
			return nil, fmt.Errorf("I_H activePush failed: %w", err)
		}
	}

	// --- Rust: `buffers_active_usable` (Min-Heap) 초기화 ---
	// 7. 초기 제약 버퍼(S+H개)를 가중치(연결된 심볼 수)와 함께 힙에 삽입합니다.
	buffersActiveUsable := newBufferWeightMap(capacity)
	for i, bufState := range bufferState {
		if bufState.activeUsedWeight == 0 {
			// Rust의 NonZeroU16::new(...).unwrap()에 해당합니다.
			// 가중치가 0인 버퍼는 힙에 추가하면 안 됩니다.
			// (G_LDPC/G_Half가 비어있는 행을 생성할 수도 있음)
			continue
		}
		buffersActiveUsable.insertBufferWeight(i, bufState.activeUsedWeight)
	}

	// 8. `buffers_inactivated`는 비어 있는 힙으로 생성합니다.
	buffersInactivated := newBufferWeightMap(capacity)

	// 9. 최종 `lowLevelDecoder` 구조체 반환
	decoder := &lowLevelDecoder{
		params:                  params,
		bufferState:             bufferState,
		intermediateSymbolState: intermediateSymbolState,
		buffersActiveUsable:     buffersActiveUsable,
		buffersInactivated:      buffersInactivated,
		numRedundantBuffers:     0,
		numSourceSymbolsPaired:  0,
	}

	// decoder.check() // (디버그용 Rust 함수, 생략)

	return decoder, nil
}

// [함수 추가] `decoder.rs`에 있던 헬퍼 함수
func (d *lowLevelDecoder) numTempBuffersRequired() int {
	// Rust: params.num_ldpc_symbols() + params.num_half_symbols()
	return int(d.params.NumLdpcSymbols) + int(d.params.NumHalfSymbols)
}

// [새 함수 추가] lowLevelDecoder의 flat index를 managedDecoder의
// bufferId로 변환하기 위한 헬퍼입니다.
// Rust의 `buffer_index_to_buffer_id`에 해당합니다.
func (d *lowLevelDecoder) bufferIndexToBufferID(flatIndex int) bufferId {
	numTemp := d.numTempBuffersRequired()
	if flatIndex < numTemp {
		// 임시 버퍼 (0..S+H-1)
		return bufferId{Type: tempBuffer, Index: flatIndex}
	}
	// 수신 버퍼 (상대 인덱스 0..)
	return bufferId{Type: receiveBuffer, Index: flatIndex - numTemp}
}

func (d *lowLevelDecoder) ReceiveSymbol(encodingSymbolID int, xorBuffers func(aFlat, bFlat int) error) error {

	// `buffer_index`는 이 새 버퍼의 *flat index*입니다.
	// (0..S+H-1은 임시 버퍼, S+H.. 이후는 수신 버퍼)
	bufferIndex := len(d.bufferState)
	bufferIndex16 := uint16(bufferIndex)

	// 이 새 버퍼에 대한 '상태' 객체 생성
	bufState := newBuffer()

	// 이 버퍼와 XOR로 연결된, *이미 해결된(Used)* 심볼들의
	// *버퍼 인덱스*를 임시로 저장할 리스트
	usedBufferIndices := make([]uint16, 0, util.MAX_DEGREE) // (Rust: Vec::with_capacity)

	// --- 1. LT 시퀀스 생성 및 버퍼 초기화 ---
	// `lt.go`의 `LTSequenceOp`를 호출하여 이 청크 ID(ESI)와
	// 연결된 모든 중간 심볼(ISIs)을 순회합니다.
	err := d.params.LTSequenceOp(encodingSymbolID, func(intermediateSymbolID int) {
		symbol := d.intermediateSymbolState[intermediateSymbolID]
		symID16 := uint16(intermediateSymbolID)

		// 새 버퍼 '상태'에 이 심볼을 추가합니다.
		// 'activeUsedWeight'는 심볼이 Inactivated가 아닐 때만 증가시킵니다.
		bufState.appendIntermediateSymbolID(symID16, !symbol.IsInactivated())

		// 만약 이 심볼이 *이미 해결된(Used)* 상태라면,
		// 해결된 값이 저장된 '버퍼 인덱스'를 리스트에 추가합니다.
		if usedBufIdx, ok := symbol.isUsedBufferIndex(); ok {
			usedBufferIndices = append(usedBufferIndices, usedBufIdx)
		}
	})
	if err != nil {
		// `lt.go`의 `rand` 또는 `deg`가 스텁이면 여기서 패닉/에러 발생
		return fmt.Errorf("LTSequenceOp failed: %w", err)
	}

	// --- 2. 버퍼 축소 (Reduction) ---
	// 이 새 버퍼(방정식)에서 이미 해결된 심볼(변수)들을
	// XOR 연산을 통해 제거합니다.

	// `managedDecoder`가 이 새 버퍼를 식별할 ID
	newBufID := d.bufferIndexToBufferID(bufferIndex)

	for _, usedBufferIndex16 := range usedBufferIndices {
		usedBufferIndex := int(usedBufferIndex16)

		// (상태 객체 XOR) 새 버퍼의 심볼 목록에서 'used' 심볼을 제거/추가
		bufState.xorEq(d.bufferState[usedBufferIndex])

		// 'Used' 심볼을 제거했으므로 'activeUsedWeight'를 1 감소시킵니다.
		bufState.activeUsedWeight-- // (Rust: `buffer.active_used_weight -= 1`)

		// (실제 바이트 XOR) `managedDecoder`에게 바이트 XOR를 요청합니다.
		// "새 버퍼의 바이트 = 새 버퍼의 바이트 ^ 이미 해결된 버퍼의 바이트"
		// `xorBuffers` 콜백은 `managedDecoder`의 `xorCallback`입니다.
		usedBufID := d.bufferIndexToBufferID(usedBufferIndex)
		err := xorBuffers(newBufID.toFlat(d.numTempBuffersRequired()), usedBufID.toFlat(d.numTempBuffersRequired()))
		if err != nil {
			return fmt.Errorf("xorBuffers callback failed: %w", err)
		}
	}

	// --- 3. 역방향 인덱싱 (Back-pointer) ---
	// 축소가 완료된 버퍼의 심볼 목록을 순회하며,
	// 각 심볼의 '상태'에 "이 새 버퍼가 너를 참조한다"고 알려줍니다.
	for _, intermediateSymbolID16 := range bufState.intermediateSymbolIDs.Values() {
		symState := d.intermediateSymbolState[intermediateSymbolID16]

		// 심볼 상태가 Active 또는 Inactivated일 때만 back-pointer 추가
		err := symState.activeInactivatedPush(bufferIndex16)
		if err != nil {
			// (Used 상태의 심볼에 push하려 할 때 발생)
			// `xorEq`가 완벽하다면 이 에러는 발생하지 않아야 합니다.
			return fmt.Errorf("activeInactivatedPush failed: %w", err)
		}
	}

	// --- 4. 새 버퍼 상태 저장 및 힙(Heap) 추가 ---
	weight := len(bufState.intermediateSymbolIDs) // 총 연결된 심볼 수
	activeUsedWeight := bufState.activeUsedWeight // '필링'에 사용될 가중치

	// 새 버퍼의 '상태' 객체를 '두뇌'의 리스트에 추가
	d.bufferState = append(d.bufferState, bufState)

	if activeUsedWeight > 0 {
		// '필링'에 사용될 수 있는 버퍼. (Weight 1순위)
		d.buffersActiveUsable.insertBufferWeight(bufferIndex, activeUsedWeight)
	} else if weight > 0 {
		// '비활성' 심볼만 포함된 버퍼. (가우스 소거법 대상)
		d.buffersInactivated.insertBufferWeight(bufferIndex, uint16(weight))
	} else {
		// (weight == 0)
		// 완전히 축소되어 아무 심볼도 남지 않은 버퍼 (중복 또는 불필요)
		d.numRedundantBuffers++
	}

	// d.check() // (Rust 디버그용, 생략)
	return nil
}

// [새 함수 추가] bufferFirstActiveIntermediateSymbol
// (Rust의 `buffer_first_active_intermediate_symbol` 포팅)
func (d *lowLevelDecoder) bufferFirstActiveIntermediateSymbol(bufferIndex uint16) (uint16, bool) {
	bufState := d.bufferState[bufferIndex]
	for _, symID := range bufState.intermediateSymbolIDs.Values() {
		if d.intermediateSymbolState[symID].IsActive() {
			return symID, true
		}
	}
	return 0, false // 찾지 못함 (로직 오류)
}

// [새 함수 추가] decrementBufferWeight
// (Rust의 `decrement_buffer_weight` 포팅)
func (d *lowLevelDecoder) decrementBufferWeight(bufferIndex uint16) {
	bufState := d.bufferState[bufferIndex]
	bufState.activeUsedWeight--

	if bufState.activeUsedWeight > 0 {
		// 가중치 갱신 (e.g., 2 -> 1)
		d.buffersActiveUsable.updateBufferWeight(int(bufferIndex), bufState.activeUsedWeight)
	} else {
		// 가중치가 0이 됨 (e.g., 1 -> 0)
		// 1. active/usable 힙에서 제거
		_, ok := d.buffersActiveUsable.removeBufferWeight(int(bufferIndex))
		if !ok {
			// 이 경우는 발생하면 안 됨
			panic(fmt.Sprintf("decrementBufferWeight: buffer %d not in active heap", bufferIndex))
		}

		// 2. 총 심볼 수(weight) 확인
		weight := len(bufState.intermediateSymbolIDs)
		if weight > 0 {
			// 비활성(inactivated) 심볼만 남음 -> 비활성 힙에 추가
			d.buffersInactivated.insertBufferWeight(int(bufferIndex), uint16(weight))
		} else {
			// 빈 버퍼가 됨 -> 중복 버퍼
			d.numRedundantBuffers++
		}
	}
}

// [새 함수 추가] bufferPeelXorEq
// (Rust의 `buffer_peel_xor_eq` 포팅)
// a = reducee (줄어들 버퍼), b = reducing (제거할 버퍼, Used 상태)
func (d *lowLevelDecoder) bufferPeelXorEq(a, b uint16) {
	aState := d.bufferState[a]
	bState := d.bufferState[b] // b는 `Used` 상태여야 함

	for _, symID := range bState.intermediateSymbolIDs.Values() {
		symbol := d.intermediateSymbolState[symID]

		if !symbol.IsInactivated() {
			// `b` 버퍼에 있는 `Active` 또는 `Used` 심볼 (Peeling 대상)
			// `a` 버퍼의 목록에서 이 심볼을 *제거*합니다.
			// (Rust: `aref.intermediate_symbol_ids.remove`)
			if !aState.intermediateSymbolIDs.Remove(symID) {
				// 이 경우는 발생하면 안 됨 (버퍼 `a`가 `b`의 심볼을 갖고 있어야 함)
				panic(fmt.Sprintf("bufferPeelXorEq: buffer %d does not contain symbol %d", a, symID))
			}
		} else {
			// `b` 버퍼에 있는 `Inactivated` 심볼
			// `a` 버퍼의 목록과 Set-XOR 연산을 수행합니다.
			// (Rust: `aref.intermediate_symbol_ids.insert_or_remove`)
			if _, found := aState.intermediateSymbolIDs.Find(symID); found {
				// (remove)
				aState.intermediateSymbolIDs.Remove(symID)
				_ = symbol.inactivatedRemove(a) // back-pointer 제거
			} else {
				// (insert)
				aState.intermediateSymbolIDs.Insert(symID)
				_ = symbol.inactivatedInsert(a) // back-pointer 추가
			}
		}
	}
}

// [새 함수 추가] tryPeel
// (Rust의 `try_peel` 포팅)
func (d *lowLevelDecoder) tryPeel(xorBuffers func(aFlat, bFlat int) error) (bool, error) {
	madeProgress := false

	for {
		if d.decodingDone() {
			break
		}

		// 1. Min-Heap에서 가중치가 가장 낮은 버퍼를 확인
		reducingBufferIndex, weight, ok := d.buffersActiveUsable.peekMin()
		if !ok || weight != 1 {
			break // 힙이 비었거나, 가중치가 1인 버퍼가 더 이상 없음
		}

		// 2. 가중치가 1인 버퍼(`reducing_buffer`)를 찾음. 힙에서 제거.
		d.buffersActiveUsable.removeMin()

		// 3. 이 버퍼가 가리키는 단 하나의 'Active' 심볼을 찾음
		intermediateSymbolID, ok := d.bufferFirstActiveIntermediateSymbol(reducingBufferIndex)
		if !ok {
			// `weight`가 1인데 Active 심볼이 없는 경우 (Used 심볼만 있음)
			// Rust 코드에서는 `used` 플래그만 설정하고 넘어가는 듯함
			// (Rust: `buffer_first_active_intermediate_symbol`이 `unwrap()` 하므로
			// 이 경우는 발생하지 않거나, `active_used_weight`가 1인 버퍼는
			// 항상 `Active` 심볼 1개를 가져야 함을 보장함)

			// Rust 코드를 다시 보니, `weight`는 `active_used_weight`임.
			// `active_used_weight`가 1이라는 것은 `Active` 심볼 1개 또는
			// `Used` 심볼 1개를 의미합니다.
			// `buffer_first_active_intermediate_symbol`은 `Active`만 찾습니다.
			// 만약 `Used` 심볼 1개만 있다면 `unwrap()`에서 패닉이 날 것입니다.
			//
			// [Rust 코드 재확인]
			// 아, `intermediate_symbol_ids`를 순회하며 `is_active()`만 찾습니다.
			// `Used` 상태는 `is_active()`가 false를 반환합니다.
			// `active_used_weight`는 `Active` *또는* `Used`의 개수입니다.
			//
			// [결론] `active_used_weight=1`일 때,
			// 1) `Active` 심볼 1개 -> `try_peel` 진행 (위의 `ok`가 true)
			// 2) `Used` 심볼 1개 -> `ok`가 false가 됨
			if !ok {
				// `active_used_weight`가 1이지만 `Active` 심볼이 없는 경우
				// (즉, `Used` 심볼 1개만 있음)
				// 이 버퍼는 이미 처리된 것(Used)으로 간주하고 루프 계속
				d.bufferState[reducingBufferIndex].used = true
				continue
			}
		}

		// 4. 버퍼와 심볼의 상태를 'Used'로 변경
		d.bufferState[reducingBufferIndex].used = true

		// `active_make_used`: 심볼 상태를 Used로 바꾸고,
		// 이 심볼을 참조하던 *다른* 버퍼들의 인덱스 목록(`reducee_buffer_indices`)을 반환받음
		reduceeBufferIndices, err := d.intermediateSymbolState[intermediateSymbolID].activeMakeUsed(reducingBufferIndex)
		if err != nil {
			return false, fmt.Errorf("activeMakeUsed failed: %w", err)
		}

		// 5. 이 심볼을 참조하던 다른 모든 버퍼(`reducee_buffer`)를 순회
		for _, reduceeBufferIndex := range reduceeBufferIndices.Values() {
			if reduceeBufferIndex == reducingBufferIndex {
				continue // 자기 자신은 제외
			}

			// 6. `reducee` 버퍼의 상태(심볼 목록) 업데이트
			// (reducee.syms = reducee.syms ^ reducing.syms)
			d.bufferPeelXorEq(reduceeBufferIndex, reducingBufferIndex)

			// 7. `reducee` 버퍼의 가중치를 1 감소시키고 힙 갱신
			d.decrementBufferWeight(reduceeBufferIndex)

			// 8. `managedDecoder`에게 실제 바이트 XOR 연산 요청
			// (reducee.bytes = reducee.bytes ^ reducing.bytes)
			err := xorBuffers(
				d.bufferIndexToBufferID(int(reduceeBufferIndex)).toFlat(d.numTempBuffersRequired()),
				d.bufferIndexToBufferID(int(reducingBufferIndex)).toFlat(d.numTempBuffersRequired()),
			)
			if err != nil {
				return false, fmt.Errorf("xorBuffers callback failed: %w", err)
			}
		}

		// 9. 복원된 심볼이 "원본 심볼"(K개 중 하나)인지 확인
		if len(d.bufferState[reducingBufferIndex].intermediateSymbolIDs) == 1 &&
			intermediateSymbolID < d.params.NumSourceSymbols {
			d.numSourceSymbolsPaired++
		}

		// d.check() // (Rust 디버그용, 생략)
		madeProgress = true
	}

	return madeProgress, nil
}

// [함수 추가] decodingDone
// (Rust의 `decoding_done` 포팅)
func (d *lowLevelDecoder) decodingDone() bool {
	return d.numSourceSymbolsPaired == int(d.params.NumSourceSymbols)
}

// [헬퍼 추가] `bufferId`를 `flatIndex`로 변환 (xorBuffers 콜백용)
func (bid bufferId) toFlat(numTempBuffers int) int {
	if bid.Type == tempBuffer {
		return bid.Index
	}
	return numTempBuffers + bid.Index
}

// [새 함수 추가] bufferInactivatedXorEq
// (Rust의 `buffer_inactivated_xor_eq` 포팅)
// `bufferPeelXorEq`와 거의 동일하지만, `Inactivated` 상태의 심볼만 처리합니다.
func (d *lowLevelDecoder) bufferInactivatedXorEq(a, b uint16) {
	aState := d.bufferState[a]
	bState := d.bufferState[b]

	for _, symID := range bState.intermediateSymbolIDs.Values() {
		symbol := d.intermediateSymbolState[symID]

		if !symbol.IsInactivated() {
			// `b` 버퍼에 있는 `Active` 또는 `Used` 심볼
			// `a` 버퍼의 목록에서 이 심볼을 *제거*합니다.
			if !aState.intermediateSymbolIDs.Remove(symID) {
				panic(fmt.Sprintf("bufferInactivatedXorEq: buffer %d does not contain symbol %d", a, symID))
			}
		} else {
			// `b` 버퍼에 있는 `Inactivated` 심볼
			// `a` 버퍼의 목록과 Set-XOR 연산을 수행합니다.
			if _, found := aState.intermediateSymbolIDs.Find(symID); found {
				// (remove)
				aState.intermediateSymbolIDs.Remove(symID)
				_ = symbol.inactivatedRemove(a) // back-pointer 제거
			} else {
				// (insert)
				aState.intermediateSymbolIDs.Insert(symID)
				_ = symbol.inactivatedInsert(a) // back-pointer 추가
			}
		}
	}
}

// [새 함수 추가] tryInactiveGaussian
// (Rust의 `try_inactive_gaussian` 포팅)
func (d *lowLevelDecoder) tryInactiveGaussian(xorBuffers func(aFlat, bFlat int) error) (bool, error) {
	// [가드 1] 비활성 힙이 비어있는지 확인
	if d.buffersInactivated.isEmpty() {
		return false, nil // 제거할 것이 없음
	}

	// [가드 2] 비활성 힙의 최소 가중치가 1인지 확인
	if _, minWeight, ok := d.buffersInactivated.peekMin(); ok && minWeight == 1 {
		return false, nil
	}

	// --- 1. 가우스 소거법 대상 수집 ---
	// (이전과 동일)
	inactivatedBufferIndices := make([]uint16, 0, len(d.buffersInactivated.heapIndexToBufferIndex))
	d.buffersInactivated.enumerate(func(bufferIndex, weight uint16) {
		inactivatedBufferIndices = append(inactivatedBufferIndices, bufferIndex)
	})
	inactivatedSymbolIDSet := make(map[uint16]struct{})
	for _, bufIdx := range inactivatedBufferIndices {
		bufState := d.bufferState[bufIdx]
		for _, symID := range bufState.intermediateSymbolIDs.Values() {
			inactivatedSymbolIDSet[symID] = struct{}{}
		}
	}
	if len(inactivatedBufferIndices) < len(inactivatedSymbolIDSet) {
		return false, nil // 풀 수 없음
	}
	inactivatedSymbolIDs := make([]uint16, 0, len(inactivatedSymbolIDSet))
	for symID := range inactivatedSymbolIDSet {
		inactivatedSymbolIDs = append(inactivatedSymbolIDs, symID)
	}
	sort.Slice(inactivatedSymbolIDs, func(i, j int) bool {
		return inactivatedSymbolIDs[i] < inactivatedSymbolIDs[j]
	})
	numRows := len(inactivatedBufferIndices)
	numCols := len(inactivatedSymbolIDs)

	// --- 2. 행렬 생성 (matrix.go) ---
	// (이전과 동일)
	mat := util.NewDenseMatrixFromFn(numRows, numCols, func(i, j int) bool {
		bufferIndex := inactivatedBufferIndices[i]
		symbolID := inactivatedSymbolIDs[j]
		_, found := d.bufferState[bufferIndex].intermediateSymbolIDs.Find(symbolID)
		return found
	})

	// --- 3. 행렬 풀이 (matrix.go) ---
	// [!!] 버그 수정: 콜백 함수의 인자 타입을 `interface{}` -> `RowOperation`으로 변경
	err := mat.RowwiseEliminationGaussianFullPivot(func(op util.RowOperation) {
		switch v := op.(type) {
		case util.RowOperationSubAssign:
			// (i, j) = (v.I, v.J)
			reduceeBufferIndex := inactivatedBufferIndices[v.I]  // 행 i
			reducingBufferIndex := inactivatedBufferIndices[v.J] // 행 j

			// 3a. '두뇌' 상태 업데이트 (Row[i] = Row[i] ^ Row[j])
			d.bufferInactivatedXorEq(reduceeBufferIndex, reducingBufferIndex)

			// 3b. '손'에게 실제 바이트 XOR 연산 요청
			_ = xorBuffers( // (콜백 에러는 일단 무시, 필요시 로깅)
				d.bufferIndexToBufferID(int(reduceeBufferIndex)).toFlat(d.numTempBuffersRequired()),
				d.bufferIndexToBufferID(int(reducingBufferIndex)).toFlat(d.numTempBuffersRequired()),
			)
		}
	})
	if err != nil {
		fmt.Printf("Warning: rowwiseEliminationGaussianFullPivot failed: %v\n", err)
		return false, nil
	}

	// --- 4. 힙(Heap) 상태 갱신 ---
	// (이전과 동일)
	for _, bufferIndex := range inactivatedBufferIndices {
		bufState := d.bufferState[bufferIndex]
		weight := len(bufState.intermediateSymbolIDs)

		if weight > 0 {
			d.buffersInactivated.updateBufferWeight(int(bufferIndex), uint16(weight))
		} else {
			d.buffersInactivated.removeBufferWeight(int(bufferIndex))
			d.numRedundantBuffers++
		}
	}

	return true, nil // 가우스 소거법 시도/성공
}

// [새 함수 추가] tryReactivateSymbols
// (Rust의 `try_reactivate_symbols` 포팅)
func (d *lowLevelDecoder) tryReactivateSymbols(xorBuffers func(aFlat, bFlat int) error) (bool, error) {
	madeProgress := false

	for {
		if d.decodingDone() {
			break
		}

		// 1. "비활성" 힙에서 가중치가 1인 버퍼를 찾습니다.
		reducingBufferIndex, weight, ok := d.buffersInactivated.peekMin()
		if !ok || weight != 1 {
			break // 힙이 비었거나 가중치가 1인 버퍼가 없음
		}

		// 2. 이 버퍼는 "Inactivated" 심볼 1개로만 구성됨. 힙에서 제거.
		d.buffersInactivated.removeMin() // (peekMin이 0번 인덱스이므로 removeMin 사용)

		reducingBuffer := d.bufferState[reducingBufferIndex]

		// 3. 그 "단 하나의" 심볼 ID를 찾습니다.
		intermediateSymbolID, ok := reducingBuffer.firstIntermediateSymbolID()
		if !ok {
			// (이론상 weight 1이므로 항상 true여야 함)
			return false, fmt.Errorf("reactivate: buffer %d has weight 1 but no symbols", reducingBufferIndex)
		}

		// 4. 버퍼와 심볼의 상태를 `Used`로 갱신합니다.
		reducingBuffer.activeUsedWeight = 1 // 이제 Active/Used 됨
		reducingBuffer.used = true

		// `inactivated_make_used`: 심볼 상태를 Used로 바꾸고,
		// 이 심볼을 참조하던 *다른* 버퍼들의 목록을 반환받음
		reduceeBufferIndices, err := d.intermediateSymbolState[intermediateSymbolID].inactivatedMakeUsed(reducingBufferIndex)
		if err != nil {
			return false, fmt.Errorf("inactivatedMakeUsed failed: %w", err)
		}

		// 5. 이 심볼을 참조하던 다른 모든 버퍼(`reducee_buffer`)를 순회
		for _, reduceeBufferIndex := range reduceeBufferIndices.Values() {
			if reduceeBufferIndex == reducingBufferIndex {
				continue // 자기 자신은 제외
			}

			reduceeBuffer := d.bufferState[reduceeBufferIndex]

			// 6. `reducee` 버퍼의 심볼 목록에서 이 심볼을 제거
			if !reduceeBuffer.intermediateSymbolIDs.Remove(intermediateSymbolID) {
				// (이론상 `inactivatedMakeUsed`가 반환했으므로 항상 true여야 함)
				panic(fmt.Sprintf("reactivate: buffer %d does not contain symbol %d", reduceeBufferIndex, intermediateSymbolID))
			}

			// 7. `reducee` 버퍼의 가중치 갱신
			// (심볼이 'Inactivated'였으므로 `activeUsedWeight`는 변경 없음)
			// `reducee` 버퍼가 `buffers_inactivated` 힙에 있는지 확인
			if reduceeBuffer.activeUsedWeight == 0 {
				weight := len(reduceeBuffer.intermediateSymbolIDs)
				if weight > 0 {
					d.buffersInactivated.updateBufferWeight(int(reduceeBufferIndex), uint16(weight))
				} else {
					// 힙에서 제거
					d.buffersInactivated.removeBufferWeight(int(reduceeBufferIndex))
					d.numRedundantBuffers++
				}
			}

			// 8. `reducee` 버퍼가 이 연산으로 인해 "해결"되었는지 확인
			// (Rust: `if reducee_buffer.is_paired() ...`)
			if reduceeBuffer.isPaired() {
				// `activeUsedWeight`가 1인 버퍼 (Peeling 대상이 됨)
				// 이 버퍼의 "단 하나" 남은 심볼이 원본 심볼(K)인지 확인
				if firstSym, ok := reduceeBuffer.firstIntermediateSymbolID(); ok {
					if firstSym < d.params.NumSourceSymbols {
						// [!!] Rust 코드의 이 로직은 버그일 수 있습니다.
						// `try_peel`과 달리, `reducee_buffer`가 `Used`가 아님에도
						// `num_source_symbols_paired`를 증가시킵니다.
						// Rust 로직을 그대로 따릅니다.
						d.numSourceSymbolsPaired++
					}
				}
			}

			// 9. `managedDecoder`에게 실제 바이트 XOR 연산 요청
			err := xorBuffers(
				d.bufferIndexToBufferID(int(reduceeBufferIndex)).toFlat(d.numTempBuffersRequired()),
				d.bufferIndexToBufferID(int(reducingBufferIndex)).toFlat(d.numTempBuffersRequired()),
			)
			if err != nil {
				return false, fmt.Errorf("xorBuffers callback failed: %w", err)
			}
		}

		// 10. *방금 해결한* 심볼이 "원본 심볼"(K개 중 하나)인지 확인
		if intermediateSymbolID < d.params.NumSourceSymbols {
			d.numSourceSymbolsPaired++
		}

		// d.check() // (Rust 디버그용, 생략)
		madeProgress = true
	}

	return madeProgress, nil
}

// [새 함수 추가] tryInactivateOneSymbol
// (Rust의 `try_inactivate_one_symbol` 포팅)
func (d *lowLevelDecoder) tryInactivateOneSymbol() bool {
	madeProgress := false

	// 1. `active` 힙에서 최소 가중치 버퍼를 확인
	bufferIndex, activeUsedWeight, ok := d.buffersActiveUsable.peekMin()
	if !ok {
		return false // `active` 힙이 비었음
	}

	// 2. 최소 가중치가 1보다 커야 함 (즉, 필링이 막힌 상태)
	if activeUsedWeight > 1 {
		madeProgress = true

		intermediateSymbolID, ok := d.bufferFirstActiveIntermediateSymbol(bufferIndex)
		if !ok {
			// (이론상 `activeUsedWeight > 1`이고 `tryReactivateSymbols`가
			//  실패했다면 `Active` 심볼이 2개 이상 있어야 함)
			panic(fmt.Sprintf("tryInactivateOneSymbol: buffer %d has weight %d but no active symbols",
				bufferIndex, activeUsedWeight))
		}

		symState := d.intermediateSymbolState[intermediateSymbolID]

		// 4. 심볼을 강제로 'Inactivated' 상태로 변경
		if err := symState.activeInactivate(); err != nil {
			panic(fmt.Sprintf("tryInactivateOneSymbol: activeInactivate failed: %v", err))
		}

		// 5. 이 심볼에 연결된 모든 버퍼를 순회
		// (Rust: `.clone()` -> Go: 슬라이스 복사 후 순회)
		inactivatedBufferIndices, err := symState.inactivatedValues()
		if err != nil {
			panic(fmt.Sprintf("tryInactivateOneSymbol: inactivatedValues failed: %v", err))
		}

		bufferIndicesToUpdate := make([]uint16, len(inactivatedBufferIndices))
		copy(bufferIndicesToUpdate, inactivatedBufferIndices)

		for _, bufIdx := range bufferIndicesToUpdate {
			// 6. 각 버퍼의 `activeUsedWeight`를 1 감소시키고 힙 갱신
			d.decrementBufferWeight(bufIdx)
		}

		// d.check() // (Rust 디버그용, 생략)
	}

	return madeProgress
}

// [함수 수정] TryDecode 메인 루프 업데이트
func (d *lowLevelDecoder) TryDecode(
	inactivationSymbolThreshold int,
	xorBuffers func(aFlat, bFlat int) error,
) (bool, error) {

	type decodingState int
	const (
		stateReactivateSymbols decodingState = iota
		statePeeling
		stateMaybeInactiveGaussian
		stateInactivateSymbol
		stateDone
	)

	usableBuffers := len(d.bufferState) - int(d.numRedundantBuffers)
	tryHarder := usableBuffers >= inactivationSymbolThreshold

	operation := stateReactivateSymbols

	for !d.decodingDone() && operation != stateDone {
		var err error
		madeProgress := false

		switch operation {
		case stateReactivateSymbols:
			madeProgress, err = d.tryReactivateSymbols(xorBuffers)
			if err != nil {
				return false, fmt.Errorf("tryReactivateSymbols failed: %w", err)
			}
			operation = statePeeling

		case statePeeling:
			madeProgress, err = d.tryPeel(xorBuffers)
			if err != nil {
				return false, fmt.Errorf("tryPeel failed: %w", err)
			}

			if madeProgress {
				operation = stateReactivateSymbols
			} else {
				operation = stateMaybeInactiveGaussian
			}

		case stateMaybeInactiveGaussian:
			if !tryHarder {
				operation = stateDone
			} else {
				madeProgress, err = d.tryInactiveGaussian(xorBuffers)
				if err != nil {
					return false, fmt.Errorf("tryInactiveGaussian failed: %w", err)
				}

				if madeProgress {
					operation = stateReactivateSymbols
				} else {
					operation = stateInactivateSymbol
				}
			}

		case stateInactivateSymbol:
			if !tryHarder {
				operation = stateDone
			} else {
				// [수정] `try_inactivate_one_symbol` 스텁을 실제 호출로 변경
				madeProgress = d.tryInactivateOneSymbol()

				if madeProgress {
					operation = stateReactivateSymbols // 성공했으면 다시 처음부터
				} else {
					operation = stateDone // 실패했으면 (더 이상 비활성화할 심볼이 없으면) 중단
				}
			}

		case stateDone:
			panic("Decoder loop entered Done state")
		}
	}

	return d.decodingDone(), nil
}

// [함수 수정] GetReconstructedOrder 스텁을 실제 구현으로 교체
// (Rust의 `source_symbol_to_buffer_id` 로직을 K번 반복하여 맵 생성)
func (d *lowLevelDecoder) GetReconstructedOrder() (map[int]bufferId, error) {
	// 1. 디코딩이 완료되었는지 확인
	if !d.decodingDone() {
		return nil, ErrDecodeNotDone
	}

	reconstructedMap := make(map[int]bufferId)
	k := int(d.params.NumSourceSymbols) // K

	// 2. 0번부터 K-1번까지 모든 "원본 심볼"을 순회
	for i := 0; i < k; i++ { // i = source_symbol_id
		symState := d.intermediateSymbolState[i]

		// 3. 해당 원본 심볼이 'Used'(해결됨) 상태인지 확인
		usedBufIdx, ok := symState.isUsedBufferIndex()
		if !ok {
			// 디코딩이 완료(decodingDone()=true)되었다면,
			// 모든 K개의 원본 심볼은 'Used' 상태여야 합니다.
			return nil, fmt.Errorf("%w: symbol %d is not in 'Used' state, though decoding is done",
				ErrReconstruction, i)
		}

		bufIdxInt := int(usedBufIdx)

		// 4. [중요] 해당 심볼이 저장된 버퍼가 "순수한지" 확인
		// (즉, 이 버퍼가 *오직* 이 심볼 하나만 포함하고 있는지 확인)
		buffer := d.bufferState[bufIdxInt]
		if len(buffer.intermediateSymbolIDs) == 1 {
			// (Rust: debug_assert!(buffer.is_paired());)

			// 5. 맵에 추가: (원본 심볼 인덱스) -> (데이터가 저장된 버퍼 ID)
			reconstructedMap[i] = d.bufferIndexToBufferID(bufIdxInt)
		} else {
			// 이 버퍼는 'Used' 상태의 심볼을 포함하지만, 다른 심볼(아마도 Inactivated)과
			// XOR된 복잡한 상태입니다.
			// (이론상 디코딩이 완료되었다면 이 상태는 발생하지 않아야 함)
			return nil, fmt.Errorf("%w: symbol %d is 'Used' in buffer %d, but buffer is not pure (len %d)",
				ErrReconstruction, i, bufIdxInt, len(buffer.intermediateSymbolIDs))
		}
	}

	// 6. K개의 심볼을 모두 찾았는지 확인
	if len(reconstructedMap) != k {
		return nil, fmt.Errorf("%w: failed to find all K=%d symbols (found %d)",
			ErrReconstruction, k, len(reconstructedMap))
	}

	return reconstructedMap, nil
}

func (d *lowLevelDecoder) nextBufferIndex() int {
	return len(d.bufferState)
}
