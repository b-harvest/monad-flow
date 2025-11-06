package common

import (
	"math/big"
	"monad-flow/model/message/protocol/no_endorsement"
	"monad-flow/util"
)

type ConsensusTip struct {
	BlockHeader *ConsensusBlockHeader
	Signature   []byte

	FreshCertificate FreshProposalCertificate `rlp:"optional"`
}

type ConsensusBlockHeader struct {
	BlockRound  util.Round
	Epoch       util.Epoch
	QC          QuorumCertificate
	Author      util.NodeID
	SeqNum      util.SeqNum
	TimestampNS big.Int

	RoundSignature          util.RoundSignature
	DelayedExecutionResults []util.FinalizedHeader
	ExecutionInputs         util.ProposedHeader
	BlockBodyID             util.ConsensusBlockBodyId

	BaseFee       *uint64 `rlp:"optional"`
	BaseFeeTrend  *uint64 `rlp:"optional"`
	BaseFeeMoment *uint64 `rlp:"optional"`
}

type NoEndorsementCertificate struct {
	Msg        *no_endorsement.NoEndorsement
	Signatures []byte
}

type FreshProposalCertificate interface {
	isFreshProposalCertificate()
}

type FreshProposalCertificateNEC struct {
	NEC *NoEndorsementCertificate
}
type FreshProposalCertificateNoTip struct {
	NoTip util.NoTipCertificate
}

func (f *FreshProposalCertificateNEC) isFreshProposalCertificate()   {}
func (f *FreshProposalCertificateNoTip) isFreshProposalCertificate() {}
