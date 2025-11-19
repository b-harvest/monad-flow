package common

import "monad-flow/util"

type BlockRange struct {
	LastBlockID util.BlockID
	NumBlocks   util.SeqNum
}
