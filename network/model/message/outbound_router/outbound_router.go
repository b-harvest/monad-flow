package outbound_router

import "github.com/ethereum/go-ethereum/rlp"

type NetworkMessageVersion struct {
	SerializeVersion   uint32
	CompressionVersion uint8
}

type OutboundRouterMessage struct {
	Version     NetworkMessageVersion
	MessageType uint8
	Message     rlp.RawValue
}
