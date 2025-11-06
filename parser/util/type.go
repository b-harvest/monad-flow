package util

import "github.com/ethereum/go-ethereum/common"

type BlockID = common.Hash
type Round uint64
type Epoch uint64
type SeqNum uint64
type ConsensusBlockBodyId = common.Hash
type NodeID []byte
type RoundSignature []byte
type ExecutionBody []byte
type ProposedHeader []byte
type FinalizedHeader []byte
type NoTipCertificate []byte
