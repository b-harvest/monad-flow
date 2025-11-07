package common

import (
	"monad-flow/util"
)

type TimeoutCertificate struct {
	Epoch      util.Epoch
	Round      util.Round
	TipRounds  []HighTipRoundSigColTuple
	HighExtend *HighExtendWrapper
}

type SignerMap struct {
	NumBits uint32
	Buf     []byte
}

type SignatureCollection struct {
	Signers SignerMap
	Sig     util.BlsAggregateSignature
}

type HighTipRoundSigColTuple struct {
	HighQCRound  util.Round
	HighTipRound util.Round
	Sigs         SignatureCollection
}
