package common

import (
	"fmt"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type HighExtend interface {
	isHighExtend()
}

type HighExtendTip struct {
	Tip           *ConsensusTip
	VoteSignature []byte
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
	TypeID uint8      `json:"typeId"`
	Extend HighExtend `json:"extend"`
}

func (w *HighExtendWrapper) DecodeRLP(s *rlp.Stream) error {
	if _, err := s.List(); err != nil {
		return fmt.Errorf("HighExtend RLP is not a list: %w", err)
	}

	typeID, err := s.Uint8()
	if err != nil {
		return fmt.Errorf("failed to decode HighExtend type ID: %w", err)
	}

	w.TypeID = typeID

	switch typeID {
	case util.HighExtendTipType:
		tipPayload := new(HighExtendTip)

		tipPayload.Tip = new(ConsensusTip)
		if err := s.Decode(tipPayload.Tip); err != nil {
			return fmt.Errorf("failed to decode HighExtendTip.Tip: %w", err)
		}

		if err := s.Decode(&tipPayload.VoteSignature); err != nil {
			if err != rlp.EOL {
				return fmt.Errorf("failed to decode HighExtendTip.VoteSignature: %w", err)
			}
		}

		w.Extend = tipPayload

	case util.HighExtendQcType:
		qcPayload := new(HighExtendQc)

		qcPayload.QC = new(QuorumCertificate)
		if err := s.Decode(qcPayload.QC); err != nil {
			return fmt.Errorf("failed to decode HighExtendQc.QC: %w", err)
		}

		w.Extend = qcPayload

	default:
		return fmt.Errorf("unknown HighExtend type ID: %d", typeID)
	}

	return s.ListEnd()
}
