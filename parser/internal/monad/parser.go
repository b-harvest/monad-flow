package monad

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"monad-flow/pkg/packet"

	"github.com/ethereum/go-ethereum/rlp"
)

var defaultProcessor = newPacketProcessor()
var (
	ErrNotEnoughSymbols     = errors.New("not enough symbols to decode the message")
	ErrInconsistentMetadata = errors.New("chunks have inconsistent AppMessageLen or Payload size")
)

// -----------------------------------------------------------------------------------

type messageReassembler struct {
	AppMessageID string
	Decoder      *raptorQDecoder
	IsComplete   bool
}

func newReassembler(appMessageID string, appMessageLen uint32, symbolSize int) (*messageReassembler, error) {
	decoder, err := newRaptorQDecoder(appMessageLen, symbolSize)
	if err != nil {
		return nil, err
	}
	return &messageReassembler{
		AppMessageID: appMessageID,
		Decoder:      decoder,
		IsComplete:   false,
	}, nil
}

// -----------------------------------------------------------------------------------

type raptorQDecoder struct {
	dataLength        uint32
	symbolSize        int
	sourceSymbolCount uint32
	symbols           map[uint16][]byte
}

func newRaptorQDecoder(dataLength uint32, symbolSize int) (*raptorQDecoder, error) {
	if symbolSize == 0 {
		return nil, errors.New("symbol size cannot be zero")
	}
	return &raptorQDecoder{
		dataLength:        dataLength,
		symbolSize:        symbolSize,
		sourceSymbolCount: (dataLength + uint32(symbolSize) - 1) / uint32(symbolSize),
		symbols:           make(map[uint16][]byte),
	}, nil
}

func (d *raptorQDecoder) addChunk(chunk *packet.MonadPacketChunk) error {
	if len(chunk.Payload) != d.symbolSize {
		if chunk.ChunkID < uint16(d.sourceSymbolCount-1) {
			return fmt.Errorf("%w: expected symbol size %d, but got %d for chunk %d",
				ErrInconsistentMetadata, d.symbolSize, len(chunk.Payload), chunk.ChunkID)
		}
	}
	if _, exists := d.symbols[chunk.ChunkID]; !exists {
		d.symbols[chunk.ChunkID] = chunk.Payload
	}
	return nil
}

func (d *raptorQDecoder) decode() ([]byte, error) {
	for i := uint32(0); i < d.sourceSymbolCount; i++ {
		if _, ok := d.symbols[uint16(i)]; !ok {
			return nil, fmt.Errorf("%w: missing source symbol with ChunkID %d", ErrNotEnoughSymbols, i)
		}
	}
	var originalMessage bytes.Buffer
	for i := uint32(0); i < d.sourceSymbolCount; i++ {
		originalMessage.Write(d.symbols[uint16(i)])
	}
	return originalMessage.Bytes()[:d.dataLength], nil
}

// -----------------------------------------------------------------------------------

type packetProcessor struct {
	reassemblers map[string]*messageReassembler
}

type decodedRouterMessage struct {
	_                  struct{} `rlp:"-"`
	SerializeVersion   uint32
	CompressionVersion uint8
	MessageType        uint8
	Payload            []byte
}

func newPacketProcessor() *packetProcessor {
	return &packetProcessor{
		reassemblers: make(map[string]*messageReassembler),
	}
}

func (p *packetProcessor) processChunk(chunk *packet.MonadPacketChunk) ([]byte, error) {
	// 1. AppMessageHash를 캐시의 키로 사용합니다.
	appMessageIDHex := fmt.Sprintf("%x", chunk.AppMessageHash)

	// 2. 캐시(Map)에서 해당 메시지의 재조합기(reassembler)를 찾거나 새로 생성합니다.
	reassembler, exists := p.reassemblers[appMessageIDHex]
	if !exists {
		// 이 메시지의 첫 번째 청크인 경우, 새로운 재조합기를 생성합니다.
		fmt.Printf("✨ [Cache] New message detected [AppMsgID: %s]. Creating reassembler.\n", appMessageIDHex)
		symbolSize := len(chunk.Payload)
		var err error
		reassembler, err = newReassembler(appMessageIDHex, chunk.AppMessageLen, symbolSize)
		if err != nil {
			return nil, err
		}
		p.reassemblers[appMessageIDHex] = reassembler
	}

	// 이미 완성된 메시지의 중복 청크는 무시합니다.
	if reassembler.IsComplete {
		return nil, nil
	}

	// 3. 디코더에 현재 청크(심볼)를 추가합니다.
	if err := reassembler.Decoder.addChunk(chunk); err != nil {
		return nil, err
	}

	// 4. RaptorQ 디코딩을 시도합니다.
	reconstructedRLP, err := reassembler.Decoder.decode()
	if err != nil {
		if errors.Is(err, ErrNotEnoughSymbols) {
			// 아직 더 많은 청크가 필요합니다.
			return nil, nil
		}
		// 다른 디코딩 오류 발생
		return nil, err
	}

	// 5. 디코딩 성공!
	reassembler.IsComplete = true
	// 처리가 끝난 메시지는 캐시에서 삭제하여 메모리를 관리합니다.
	defer delete(p.reassemblers, appMessageIDHex)

	// 6. 최종 RLP 역직렬화를 통해 실제 페이로드를 추출하고 반환합니다.
	var routerMsg decodedRouterMessage
	err = rlp.Decode(bytes.NewReader(reconstructedRLP), &[]interface{}{
		&routerMsg.SerializeVersion,
		&routerMsg.CompressionVersion,
		&routerMsg.MessageType,
		&routerMsg.Payload,
	})
	if err != nil {
		return nil, fmt.Errorf("final rlp decoding failed: %w", err)
	}

	return routerMsg.Payload, nil
}

func (p *packetProcessor) processPacket(data []byte) {
	packet, err := parseMonadPacket(data)
	if err != nil {
		fmt.Println("chunk parsing failed: %w", err)
		return
	}
	// debug.PrintMonadPacketDetails(packet)
	finalPayload, err := p.processChunk(packet)

	if err != nil {
		fmt.Printf("‼️  Error processing chunk: %v\n", err)
		return
	}

	// 5. 결과를 확인합니다.
	if finalPayload != nil {
		// 메시지가 성공적으로 재조합되었습니다!
		fmt.Println("✅ MESSAGE RECONSTRUCTED!")
		fmt.Printf("   Final Decoded Payload(First 10B): %x...\n", finalPayload[:min(10, len(finalPayload))])
	} else {
		// 아직 더 많은 패킷이 필요합니다.
		fmt.Println("⏳ Message incomplete, waiting for more chunks...")
	}
}

// -----------------------------------------------------------------------------------

func HexDumpAndReassemble(data []byte) {
	fmt.Printf("\n---[ Packet Received by Parser (%d bytes) ]---\n", len(data))
	defaultProcessor.processPacket(data)
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
