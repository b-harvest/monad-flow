package proposal

import (
	"monad-flow/model/message/protocol/common"
	"monad-flow/util"
)

type ProposalMessage struct {
	ProposalRound util.Round
	ProposalEpoch util.Epoch
	Tip           *common.ConsensusTip
	BlockBody     *ConsensusBlockBody
	LastRoundTC   *common.TimeoutCertificate `rlp:"optional"`
}

type ConsensusBlockBody struct {
	ExecutionBody util.ExecutionBody
}

func (*ProposalMessage) IsProtocolMessage() {}
