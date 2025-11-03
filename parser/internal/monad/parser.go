package monad

import (
	"encoding/binary"
	"fmt"
	"monad-flow/pkg/debug"
	"monad-flow/pkg/packet"
)

var defaultProcessor = newPacketProcessor()

type raptorQDecoder struct {
	dataLength   uint32
	symbols      map[uint16][]byte
	sourceSymNum uint32
}

type messageReassembler struct {
	AppMessageID string
	Decoder      *raptorQDecoder
	IsComplete   bool
}

type packetProcessor struct {
	reassemblers map[string]*messageReassembler
}

func newPacketProcessor() *packetProcessor {
	return &packetProcessor{
		reassemblers: make(map[string]*messageReassembler),
	}
}

func newRaptorQDecoder(dataLength uint32, symbolSize int) *raptorQDecoder {
	return &raptorQDecoder{
		dataLength:   dataLength,
		symbols:      make(map[uint16][]byte),
		sourceSymNum: (dataLength + uint32(symbolSize) - 1) / uint32(symbolSize),
	}
}

// -----------------------------------------------------------------------------------

func HexDumpAndReassemble(data []byte) {
	fmt.Printf("\n---[ Packet Received by Parser (%d bytes) ]---\n", len(data))
	defaultProcessor.processPacket(data)
}

func (p *packetProcessor) processPacket(data []byte) {
	packet, err := parseMonadPacket(data)
	if err != nil {
		fmt.Println("chunk parsing failed: %w", err)
		return
	}
	debug.PrintMonadPacketDetails(packet)
	appMessageHash := fmt.Sprintf("%x", packet.AppMessageHash)
	reassembler, exists := p.reassemblers[appMessageHash]

	if !exists {
		payloadSize := len(packet.Payload)
		if payloadSize == 0 {
			fmt.Println("received a chunk with zero-length symbol data")
			return
		}
		reassembler = &messageReassembler{
			AppMessageID: appMessageHash,
			Decoder:      newRaptorQDecoder(packet.AppMessageLen, payloadSize),
		}
		p.reassemblers[appMessageHash] = reassembler
	}
}

func parseMonadPacket(data []byte) (*packet.MonadPacketChunk, error) {
	if len(data) < 70 {
		return nil, fmt.Errorf("data too short: expected at least 70 bytes, got %d", len(data))
	}

	packet := &packet.MonadPacketChunk{}
	offset := 0

	// 1. Signature (65 bytes)
	copy(packet.Signature[:], data[offset:offset+65])
	offset += 65

	// 2. Version (2 bytes)
	if offset+2 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing Version")
	}
	packet.Version = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// 3. Flags (1 byte)
	if offset+1 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing Flags")
	}
	packet.Flags = data[offset]
	packet.MerkleTreeDepth = packet.Flags & 0x0F
	offset += 1

	// 4. Epoch # (8 bytes u64)
	if offset+8 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing Epoch")
	}
	packet.Epoch = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// 5. Unix timestamp in milliseconds (8 bytes u64)
	if offset+8 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing TimestampMs")
	}
	packet.TimestampMs = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// 6. first 20 bytes of hash of AppMessage (20 bytes)
	if offset+20 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing AppMessageHash")
	}
	copy(packet.AppMessageHash[:], data[offset:offset+20])
	offset += 20

	// 7. Serialized AppMessage length (4 bytes u32)
	if offset+4 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing AppMessageLen")
	}
	packet.AppMessageLen = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// 8. Merkle proof (20 bytes * (merkle_tree_depth - 1))
	if packet.MerkleTreeDepth > 0 {
		proofSize := int(packet.MerkleTreeDepth-1) * 20
		if offset+proofSize > len(data) {
			return nil, fmt.Errorf("unexpected end of data while parsing MerkleProof: expected %d bytes, only %d remaining", proofSize, len(data)-offset)
		}
		packet.MerkleProof = make([][]byte, 0, packet.MerkleTreeDepth-1)
		for i := 0; i < int(packet.MerkleTreeDepth-1); i++ {
			proofHash := make([]byte, 20)
			copy(proofHash, data[offset:offset+20])
			packet.MerkleProof = append(packet.MerkleProof, proofHash)
			offset += 20
		}
	} else {
		packet.MerkleProof = [][]byte{}
	}

	// 9. first 20 bytes of hash of chunk's first hop recipient (20 bytes)
	if offset+20 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing FirstHopRecipient")
	}
	copy(packet.FirstHopRecipient[:], data[offset:offset+20])
	offset += 20

	// 10. Chunk's merkle leaf idx (1 byte)
	if offset+1 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing MerkleLeafIdx")
	}
	packet.MerkleLeafIdx = data[offset]
	offset += 1

	// 11. reserved (1 byte)
	if offset+1 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing Reserved byte")
	}
	packet.Reserved = data[offset]
	offset += 1

	// 12. This chunk's id (2 bytes u16)
	if offset+2 > len(data) {
		return nil, fmt.Errorf("unexpected end of data while parsing ChunkID")
	}
	packet.ChunkID = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// 13. Payload (rest of data)
	packet.Payload = data[offset:]

	return packet, nil
}
