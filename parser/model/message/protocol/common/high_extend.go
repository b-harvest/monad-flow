package common

import (
	"bytes"
	"fmt"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type HighExtend interface {
	isHighExtend()
}

type HighExtendTip struct {
	Tip           *ConsensusTip
	VoteSignature []byte `rlp:"optional"`
}

func (h *HighExtendTip) isHighExtend() {}

func (h *HighExtendTip) DecodeRLP(s *rlp.Stream) error {
	return s.Decode(&h.Tip)
}

type HighExtendQc struct {
	QC *QuorumCertificate
}

func (h *HighExtendQc) isHighExtend() {}

func (h *HighExtendQc) DecodeRLP(s *rlp.Stream) error {
	return s.Decode(&h.QC)
}

type HighExtendWrapper struct {
	Extend HighExtend
}

func (w *HighExtendWrapper) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return fmt.Errorf("failed to get raw RLP for HighExtend: %w", err)
	}

	if len(raw) == 0 || (len(raw) == 1 && raw[0] == 0x80) {
		w.Extend = nil
		return nil
	}

	s = rlp.NewStream(bytes.NewReader(raw), uint64(len(raw)))

	if _, err := s.List(); err != nil {
		return fmt.Errorf("HighExtend RLP is not a list: %w", err)
	}

	typeID, err := s.Uint8()
	if err != nil {
		return fmt.Errorf("failed to decode HighExtend type ID: %w", err)
	}

	var payload HighExtend
	switch typeID {
	case util.HighExtendTipType:
		payload = new(HighExtendTip)
	case util.HighExtendQcType:
		payload = new(HighExtendQc)
	default:
		return fmt.Errorf("unknown HighExtend type ID: %d", typeID)
	}

	if err := s.Decode(payload); err != nil {
		return fmt.Errorf("failed to decode HighExtend payload type %d: %w", typeID, err)
	}

	w.Extend = payload
	return s.ListEnd()
}
