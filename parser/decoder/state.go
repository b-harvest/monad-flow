package decoder

import (
	"errors"
	"monad-flow/util"
)

type symbolState int

const (
	symbolStateActive symbolState = iota
	symbolStateInactivated
	symbolStateUsed
)

type intermediateSymbol struct {
	state symbolState

	bufferIndices util.OrderedSet

	bufferIndexUsed uint16
}

func newIntermediateSymbol() *intermediateSymbol {
	return &intermediateSymbol{
		state:         symbolStateActive,
		bufferIndices: util.NewOrderedSet(),
	}
}

func (is *intermediateSymbol) IsActive() bool {
	return is.state == symbolStateActive
}

func (is *intermediateSymbol) IsInactivated() bool {
	return is.state == symbolStateInactivated
}

func (is *intermediateSymbol) IsUsed() bool {
	return is.state == symbolStateUsed
}

func (is *intermediateSymbol) isUsedBufferIndex() (uint16, bool) {
	if is.state == symbolStateUsed {
		return is.bufferIndexUsed, true
	}
	return 0, false
}

func (is *intermediateSymbol) activeMakeUsed(bufferIndex uint16) (util.OrderedSet, error) {
	if !is.IsActive() {
		return nil, errors.New("activeMakeUsed called on non-Active symbol")
	}

	oldIndices := is.bufferIndices

	is.state = symbolStateUsed
	is.bufferIndexUsed = bufferIndex
	is.bufferIndices = nil

	return oldIndices, nil
}

func (is *intermediateSymbol) activeInactivate() error {
	if !is.IsActive() {
		return errors.New("activeInactivate called on non-Active symbol")
	}
	is.state = symbolStateInactivated
	return nil
}

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

func (is *intermediateSymbol) activePush(bufferIndex uint16) error {
	if !is.IsActive() {
		return errors.New("activePush called on non-Active symbol")
	}
	is.bufferIndices.Append(bufferIndex)
	return nil
}

func (is *intermediateSymbol) activeInactivatedPush(bufferIndex uint16) error {
	if is.IsActive() || is.IsInactivated() {
		is.bufferIndices.Append(bufferIndex)
		return nil
	}
	return errors.New("activeInactivatedPush called on Used symbol")
}

func (is *intermediateSymbol) inactivatedValues() (util.OrderedSet, error) {
	if !is.IsInactivated() {
		return nil, errors.New("inactivatedValues called on non-Inactivated symbol")
	}
	return is.bufferIndices, nil
}

func (is *intermediateSymbol) inactivatedInsert(bufferIndex uint16) error {
	if !is.IsInactivated() {
		return errors.New("inactivatedInsert called on non-Inactivated symbol")
	}
	is.bufferIndices.Insert(bufferIndex)
	return nil
}

func (is *intermediateSymbol) inactivatedRemove(bufferIndex uint16) error {
	if !is.IsInactivated() {
		return errors.New("inactivatedRemove called on non-Inactivated symbol")
	}
	is.bufferIndices.Remove(bufferIndex)
	return nil
}
