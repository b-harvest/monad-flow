package packet

import (
	"github.com/google/gopacket/layers"
)

type MonadChunkPacket struct {
	// 헤더 (Signature 이후)
	Signature       [65]byte // Signature of sender
	Version         uint16   // 2 bytes
	Flags           byte     // 1 byte (broadcast, secondary_broadcast, unused)
	MerkleTreeDepth byte     // 4 bits from Flags byte
	Epoch           uint64   // 8 bytes (u64)
	TimestampMs     uint64   // 8 bytes (u64)
	AppMessageHash  [20]byte // 20 bytes (first 20 bytes of hash of AppMessage)
	AppMessageLen   uint32   // 4 bytes (u32)
	MerkleProof     [][]byte // 20 bytes * (MerkleTreeDepth - 1)

	// 청크 특정 정보
	FirstHopRecipient [20]byte // 20 bytes (first 20 bytes of hash of chunk's first hop recipient)
	MerkleLeafIdx     byte     // 1 byte
	Reserved          byte     // 1 byte
	ChunkID           uint16   // 2 bytes (u16)
	Payload           []byte   // rest of data
}

type Packet struct {
	EthernetLayer *layers.Ethernet
	IPv4Layer     *layers.IPv4
	TCPLayer      *layers.TCP
	UDPLayer      *layers.UDP
	Payload       []byte
}
