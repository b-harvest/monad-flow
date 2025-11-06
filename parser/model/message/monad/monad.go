package monad

import (
	"github.com/ethereum/go-ethereum/rlp"
)

type MonadVersion struct {
	ProtocolVersion    uint32
	ClientVersionMajor uint16
	ClientVersionMinor uint16
	HashVersion        uint16
	SerializeVersion   uint16
}

type MonadMessage struct {
	Version MonadVersion
	TypeID  uint8
	Payload rlp.RawValue
}
