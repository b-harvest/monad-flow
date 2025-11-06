package common

import "monad-flow/util"

type TimeoutCertificate struct {
	Epoch      util.Epoch
	Round      util.Round
	TipRounds  []HighTipRoundSigColTuple
	HighExtend HighExtend
}

type HighTipRoundSigColTuple struct {
	HighQCRound  util.Round
	HighTipRound util.Round
	Sigs         []byte
}

type HighExtend interface {
	isHighExtend()
}

type HighExtendTip struct {
	Tip *ConsensusTip
}
type HighExtendQc struct {
	QC *QuorumCertificate
}

func (h *HighExtendTip) isHighExtend() {}
func (h *HighExtendQc) isHighExtend()  {}
