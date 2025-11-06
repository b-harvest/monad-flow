package timeout

import (
	"monad-flow/model/message/protocol/common"
	"monad-flow/util"
)

type TimeoutMessage struct {
	TMInfo               *TimeoutInfo
	TimeoutSignature     []byte
	HighExtend           HighExtendVote
	LastRoundCertificate common.RoundCertificate `rlp:"optional"`
}

type TimeoutInfo struct {
	Epoch        util.Epoch
	Round        util.Round
	HighQCRound  util.Round
	HighTipRound util.Round
}

type HighExtendVote interface {
	isHighExtendVote()
}

type HighExtendVoteTip struct {
	Tip *common.ConsensusTip
	Sig []byte `rlp:"optional"` // Tip(tip, Option<Sig>)
}
type HighExtendVoteQc struct {
	QC *common.QuorumCertificate
}

func (h *HighExtendVoteTip) isHighExtendVote() {}
func (h *HighExtendVoteQc) isHighExtendVote()  {}

func (*TimeoutMessage) IsProtocolMessage() {}
