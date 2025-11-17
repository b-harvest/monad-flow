package common

import (
	"monad-flow/model/message/protocol/vote"

	"github.com/ethereum/go-ethereum/rlp"
)

type QuorumCertificate struct {
	Info       vote.Vote
	Signatures rlp.RawValue
}
