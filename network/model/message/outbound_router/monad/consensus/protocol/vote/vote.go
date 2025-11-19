package vote

import "monad-flow/util"

type Vote struct {
	ID    util.BlockID
	Round util.Round
	Epoch util.Epoch
}

type VoteMessage struct {
	Vote Vote
	Sig  []byte
}
