package timeout

import (
	"monad-flow/model/message/protocol/common"
	"monad-flow/util"
)

type TimeoutMessage struct {
	TMInfo               *TimeoutInfo
	TimeoutSignature     []byte
	HighExtend           common.HighExtendWrapper
	LastRoundCertificate *common.RoundCertificateWrapper `rlp:"optional"`
}

type TimeoutInfo struct {
	Epoch        util.Epoch
	Round        util.Round
	HighQCRound  util.Round
	HighTipRound util.Round
}

func (*TimeoutMessage) IsProtocolMessage() {}
