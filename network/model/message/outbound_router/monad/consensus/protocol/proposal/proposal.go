package proposal

import (
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/common"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/core/types"
)

type Ommer struct{}

type ExecutionBody struct {
	Transactions []*types.Transaction
	Ommers       []*Ommer
	Withdrawals  []*types.Withdrawal
}

type ProposalMessage struct {
	ProposalRound util.Round
	ProposalEpoch util.Epoch
	Tip           *common.ConsensusTip
	BlockBody     *ConsensusBlockBody
	LastRoundTC   *common.TimeoutCertificate `rlp:"optional"`
}

type ConsensusBlockBody struct {
	ExecutionBody ExecutionBody
}
