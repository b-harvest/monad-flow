package decoder

import (
	"errors"
	"monad-flow/util"
)

// --- intermediateSymbol ---

// symbolState는 IntermediateSymbol의 3가지 상태를 정의합니다.
// (Rust의 enum variant)
type symbolState int

const (
	symbolStateActive symbolState = iota
	symbolStateInactivated
	symbolStateUsed
)

// intermediateSymbol은 '두뇌'가 해결하려는 단일 변수(중간 심볼)의 상태입니다.
// Rust의 `IntermediateSymbol` enum을 Go struct로 포팅한 것입니다.
type intermediateSymbol struct {
	state symbolState

	// Active 또는 Inactivated 상태일 때 사용
	// (두 상태가 동시에 사용되지 않으므로 하나의 필드를 공유)
	bufferIndices util.OrderedSet

	// Used 상태일 때 사용
	bufferIndexUsed uint16
}

// newIntermediateSymbol은 Active 상태의 새 심볼을 생성합니다.
// Rust의 `IntermediateSymbol::new()`에 해당합니다.
func newIntermediateSymbol() *intermediateSymbol {
	return &intermediateSymbol{
		state:         symbolStateActive,
		bufferIndices: util.NewOrderedSet(),
	}
}

// --- 상태 확인 메소드 (State Check Methods) ---

func (is *intermediateSymbol) IsActive() bool {
	return is.state == symbolStateActive
}

func (is *intermediateSymbol) IsInactivated() bool {
	return is.state == symbolStateInactivated
}

func (is *intermediateSymbol) IsUsed() bool {
	return is.state == symbolStateUsed
}

// isUsedBufferIndex는 심볼이 'Used' 상태인 경우
// 연결된 버퍼 인덱스를 반환합니다. (Rust의 is_used_buffer_index)
func (is *intermediateSymbol) isUsedBufferIndex() (uint16, bool) {
	if is.state == symbolStateUsed {
		return is.bufferIndexUsed, true
	}
	return 0, false
}

// --- 상태 변경 메소드 (State Transition Methods) ---

// activeMakeUsed는 Active -> Used로 상태를 변경하고,
// 기존의 bufferIndices를 반환합니다.
func (is *intermediateSymbol) activeMakeUsed(bufferIndex uint16) (util.OrderedSet, error) {
	if !is.IsActive() {
		return nil, errors.New("activeMakeUsed called on non-Active symbol")
	}

	oldIndices := is.bufferIndices

	is.state = symbolStateUsed
	is.bufferIndexUsed = bufferIndex
	is.bufferIndices = nil // 더 이상 사용하지 않음

	return oldIndices, nil
}

// activeInactivate는 Active -> Inactivated로 상태를 변경합니다.
func (is *intermediateSymbol) activeInactivate() error {
	if !is.IsActive() {
		return errors.New("activeInactivate called on non-Active symbol")
	}
	// bufferIndices 필드는 Inactivated 상태에서도 계속 사용됨
	is.state = symbolStateInactivated
	return nil
}

// inactivatedMakeUsed는 Inactivated -> Used로 상태를 변경하고,
// 기존의 bufferIndices를 반환합니다.
func (is *intermediateSymbol) inactivatedMakeUsed(bufferIndex uint16) (util.OrderedSet, error) {
	if !is.IsInactivated() {
		return nil, errors.New("inactivatedMakeUsed called on non-Inactivated symbol")
	}

	oldIndices := is.bufferIndices

	is.state = symbolStateUsed
	is.bufferIndexUsed = bufferIndex
	is.bufferIndices = nil

	return oldIndices, nil
}

// --- 상태별 동작 메소드 (State-Specific Methods) ---

// activePush는 Active 상태의 심볼에 버퍼 인덱스를 추가합니다.
// (Rust의 active_push)
func (is *intermediateSymbol) activePush(bufferIndex uint16) error {
	if !is.IsActive() {
		return errors.New("activePush called on non-Active symbol")
	}
	is.bufferIndices.Append(bufferIndex)
	return nil
}

// activeInactivatedPush는 Active 또는 Inactivated 상태의 심볼에 버퍼 인덱스를 추가합니다.
// (Rust의 active_inactivated_push)
func (is *intermediateSymbol) activeInactivatedPush(bufferIndex uint16) error {
	if is.IsActive() || is.IsInactivated() {
		is.bufferIndices.Append(bufferIndex)
		return nil
	}
	return errors.New("activeInactivatedPush called on Used symbol")
}

// inactivatedValues는 Inactivated 상태의 버퍼 목록을 반환합니다.
func (is *intermediateSymbol) inactivatedValues() (util.OrderedSet, error) {
	if !is.IsInactivated() {
		return nil, errors.New("inactivatedValues called on non-Inactivated symbol")
	}
	return is.bufferIndices, nil
}

// inactivatedInsert는 Inactivated 상태의 심볼에 버퍼 인덱스를 삽입합니다.
func (is *intermediateSymbol) inactivatedInsert(bufferIndex uint16) error {
	if !is.IsInactivated() {
		return errors.New("inactivatedInsert called on non-Inactivated symbol")
	}
	is.bufferIndices.Insert(bufferIndex)
	return nil
}

// inactivatedRemove는 Inactivated 상태의 심볼에서 버퍼 인덱스를 제거합니다.
func (is *intermediateSymbol) inactivatedRemove(bufferIndex uint16) error {
	if !is.IsInactivated() {
		return errors.New("inactivatedRemove called on non-Inactivated symbol")
	}
	is.bufferIndices.Remove(bufferIndex)
	return nil
}
