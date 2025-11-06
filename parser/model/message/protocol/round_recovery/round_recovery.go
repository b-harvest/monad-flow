package round_recovery

import (
	"monad-flow/model/message/protocol/common"
	"monad-flow/util"
)

type RoundRecoveryMessage struct {
	Round util.Round
	Epoch util.Epoch
	TC    *common.TimeoutCertificate
}

func (*RoundRecoveryMessage) IsProtocolMessage() {}
