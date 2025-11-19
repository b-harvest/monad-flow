package advanced_round

import (
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/common"

	"github.com/ethereum/go-ethereum/rlp"
)

type AdvanceRoundMessage struct {
	LastRoundCertificate *common.RoundCertificateWrapper
}

func (a *AdvanceRoundMessage) DecodeRLP(s *rlp.Stream) error {
	if _, err := s.List(); err != nil {
		return err
	}

	a.LastRoundCertificate = new(common.RoundCertificateWrapper)
	return s.Decode(a.LastRoundCertificate)
}
