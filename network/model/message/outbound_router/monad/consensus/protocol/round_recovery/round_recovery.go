package round_recovery

import (
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/common"
	"monad-flow/util"
)

type RoundRecoveryMessage struct {
	Round util.Round
	Epoch util.Epoch
	TC    *common.TimeoutCertificate
}
