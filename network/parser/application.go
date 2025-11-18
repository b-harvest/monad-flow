package parser

import (
	"encoding/binary"
	"fmt"
	"monad-flow/model"
	"sync"

	"github.com/zishang520/socket.io/clients/socket/v3"
)

func ParseMonadChunkPacket(packet model.Packet, data []byte, client *socket.Socket, clientMutex *sync.Mutex) (*model.MonadChunkPacket, error) {
	chunk := &model.MonadChunkPacket{}

	network := model.MonadNetworkPacket{}
	network.Ipv4.SrcIp = packet.IPv4Layer.SrcIP.String()
	network.Ipv4.DstIp = packet.IPv4Layer.DstIP.String()
	network.Ipv4.Protocol = packet.IPv4Layer.Protocol.String()
	if packet.TCPLayer != nil {
		network.Port.SrcPort = int(packet.TCPLayer.SrcPort)
		network.Port.DstPort = int(packet.TCPLayer.DstPort)
	} else if packet.UDPLayer != nil {
		network.Port.SrcPort = int(packet.UDPLayer.SrcPort)
		network.Port.DstPort = int(packet.UDPLayer.DstPort)
	}
	chunk.Network = network

	offset := 0

	// 1. Signature (65 bytes)
	copy(chunk.Signature[:], data[offset:offset+65])
	offset += 65

	// 2. Version (2 bytes)
	if offset+2 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing Version")
	}
	chunk.Version = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// 3. Flags (1 byte)
	if offset+1 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing Flags")
	}
	chunk.Flags = data[offset]
	chunk.Broadcast = (chunk.Flags>>7)&0x01 == 1
	chunk.SecondaryBroadcast = (chunk.Flags>>6)&0x01 == 1
	chunk.MerkleTreeDepth = chunk.Flags & 0x0F
	offset += 1

	// 4. Epoch # (8 bytes u64)
	if offset+8 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing Epoch")
	}
	chunk.Epoch = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// 5. Unix timestamp in milliseconds (8 bytes u64)
	if offset+8 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing TimestampMs")
	}
	chunk.TimestampMs = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// 6. first 20 bytes of hash of AppMessage (20 bytes)
	if offset+20 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing AppMessageHash")
	}
	copy(chunk.AppMessageHash[:], data[offset:offset+20])
	offset += 20

	// 7. Serialized AppMessage length (4 bytes u32)
	if offset+4 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing AppMessageLen")
	}
	chunk.AppMessageLen = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// 8. Merkle proof (20 bytes * (merkle_tree_depth - 1))
	if chunk.MerkleTreeDepth > 0 {
		proofSize := int(chunk.MerkleTreeDepth-1) * 20
		if offset+proofSize > len(data) {
			return nil, fmt.Errorf("unexpected end of data while parsing MerkleProof: expected %d bytes, only %d remaining", proofSize, len(data)-offset)
		}
		chunk.MerkleProof = make([][]byte, 0, chunk.MerkleTreeDepth-1)
		for i := 0; i < int(chunk.MerkleTreeDepth-1); i++ {
			proofHash := make([]byte, 20)
			copy(proofHash, data[offset:offset+20])
			chunk.MerkleProof = append(chunk.MerkleProof, proofHash)
			offset += 20
		}
	} else {
		chunk.MerkleProof = [][]byte{}
	}

	// 9. first 20 bytes of hash of chunk's first hop recipient (20 bytes)
	if offset+20 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing FirstHopRecipient")
	}
	copy(chunk.FirstHopRecipient[:], data[offset:offset+20])
	offset += 20

	// 10. Chunk's merkle leaf idx (1 byte)
	if offset+1 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing MerkleLeafIdx")
	}
	chunk.MerkleLeafIdx = data[offset]
	offset += 1

	// 11. reserved (1 byte)
	if offset+1 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing Reserved byte")
	}
	chunk.Reserved = data[offset]
	offset += 1

	// 12. This chunk's id (2 bytes u16)
	if offset+2 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing ChunkID")
	}
	chunk.ChunkID = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// 13. Payload (rest of data)
	chunk.Payload = data[offset:]

	return chunk, nil
}
