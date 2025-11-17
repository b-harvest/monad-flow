package util

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type BlockID = common.Hash
type Round uint64
type Epoch uint64
type SeqNum uint64
type ConsensusBlockBodyId = common.Hash
type NodeID []byte
type RoundSignature []byte
type Signature []byte
type FinalizedHeader types.Header
type BlsAggregateSignature []byte
