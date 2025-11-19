package no_endorsement

import "monad-flow/util"

type NoEndorsement struct {
	Epoch      util.Epoch
	Round      util.Round
	TipQCRound util.Round
}

type NoEndorsementMessage struct {
	Msg *NoEndorsement
	Sig []byte
}
