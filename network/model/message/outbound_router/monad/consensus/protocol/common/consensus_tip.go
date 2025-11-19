package common

import (
	"errors"
	"fmt"
	"math/big"
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/no_endorsement"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type ProposedHeader struct {
	OmmersHash            common.Hash
	Beneficiary           common.Address
	TransactionsRoot      common.Hash
	Difficulty            uint64
	Number                uint64
	GasLimit              uint64
	Timestamp             uint64
	ExtraData             [32]byte
	MixHash               common.Hash
	Nonce                 [8]byte
	BaseFeePerGas         uint64
	WithdrawalsRoot       common.Hash
	BlobGasUsed           uint64
	ExcessBlobGas         uint64
	ParentBeaconBlockRoot common.Hash
	RequestsHash          *common.Hash `rlp:"optional"`
}

type ConsensusTip struct {
	BlockHeader      *ConsensusBlockHeader
	Signature        []byte
	FreshCertificate *FreshProposalCertificateWrapper
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
	ExecutionInputs         ProposedHeader
	BlockBodyID             util.ConsensusBlockBodyId

	BaseFee       *uint64 `rlp:"optional"`
	BaseFeeTrend  *uint64 `rlp:"optional"`
	BaseFeeMoment *uint64 `rlp:"optional"`
}

type NoEndorsementCertificate struct {
	Msg        *no_endorsement.NoEndorsement
	Signatures []byte
}

func (ct *ConsensusTip) DecodeRLP(s *rlp.Stream) error {
	if _, err := s.List(); err != nil {
		return fmt.Errorf("ConsensusTip RLP is not a list: %w", err)
	}

	if err := s.Decode(&ct.BlockHeader); err != nil {
		return fmt.Errorf("failed to decode ConsensusTip.BlockHeader: %w", err)
	}

	if err := s.Decode(&ct.Signature); err != nil {
		return fmt.Errorf("failed to decode ConsensusTip.Signature: %w", err)
	}

	var wrapper FreshProposalCertificateWrapper
	err := s.Decode(&wrapper)

	switch {
	case err == nil:
		ct.FreshCertificate = &wrapper
	case errors.Is(err, rlp.EOL):
		ct.FreshCertificate = nil
	default:
		return fmt.Errorf("failed to decode ConsensusTip.FreshCertificate: %w", err)
	}

	return s.ListEnd()
}
