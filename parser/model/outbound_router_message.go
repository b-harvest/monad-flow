package model

type NetworkMessageVersion struct {
	SerializeVersion   uint32
	CompressionVersion uint8
}

type OutboundRouterMessage struct {
	Version     NetworkMessageVersion
	MessageType uint8
	Message     MonadMessage
}
