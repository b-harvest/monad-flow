package common

import (
	"bytes"
	"fmt"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type NoTipCertificate struct {
	Epoch     util.Epoch
	Round     util.Round
	TipRounds []HighTipRoundSigColTuple
	HighQc    *QuorumCertificate
}

type FreshProposalCertificateWrapper struct {
	Certificate FreshProposalCertificate
}

type FreshProposalCertificate interface {
	isFreshProposalCertificate()
}

type FreshProposalCertificateNEC struct {
	NEC *NoEndorsementCertificate
}

type FreshProposalCertificateNoTip struct {
	NoTip *NoTipCertificate
}

func (f *FreshProposalCertificateNEC) isFreshProposalCertificate()   {}
func (f *FreshProposalCertificateNoTip) isFreshProposalCertificate() {}

func (w *FreshProposalCertificateWrapper) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()

	if err != nil {
		return fmt.Errorf("failed to get raw RLP for FreshProposalCertificate: %w", err)
	}

	if len(raw) == 1 && raw[0] == 0x80 {
		w.Certificate = nil
		return nil
	}

	s = rlp.NewStream(bytes.NewReader(raw), uint64(len(raw)))

	// 1. RLP 리스트 시작
	_, err = s.List()
	if err != nil {
		return fmt.Errorf("FreshProposalCertificate RLP is not a list (and not nil): %w", err)
	}

	// 2. 타입 ID (u8) 디코딩
	typeID, err := s.Uint8()
	if err != nil {
		return fmt.Errorf("failed to decode FreshProposalCertificate type ID: %w", err)
	}

	// 3. 타입 ID에 따라 실제 페이로드 디코딩
	switch typeID {
	case util.NEC:
		var nec NoEndorsementCertificate
		if err := s.Decode(&nec); err != nil {
			return fmt.Errorf("failed to decode NoEndorsementCertificate payload (type 1): %w", err)
		}
		w.Certificate = &FreshProposalCertificateNEC{NEC: &nec}

	case util.NoTip:
		var noTip NoTipCertificate
		if err := s.Decode(&noTip); err != nil {
			return fmt.Errorf("failed to decode NoTipCertificate payload (type 2): %w", err)
		}
		w.Certificate = &FreshProposalCertificateNoTip{NoTip: &noTip}

	default:
		return fmt.Errorf("unknown FreshProposalCertificate type ID: %d", typeID)
	}

	// 4. RLP 리스트 종료
	return s.ListEnd()
}
