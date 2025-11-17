package model

import (
	"fmt"

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

func (packet *MonadChunkPacket) PrintMonadPacketDetails() {
	isBroadcast := (packet.Flags>>7)&0x01 == 1
	isSecondaryBroadcast := (packet.Flags>>6)&0x01 == 1

	fmt.Println("--- Monad Packet Chunk Details ---")

	// 헤더 정보
	fmt.Printf("📦 **Packet Header**\n")
	fmt.Printf("  - Signature:           %x...\n", packet.Signature[:5])
	fmt.Printf("  - Version:             %d\n", packet.Version)
	fmt.Printf("  - Flags (Raw):         0x%02x\n", packet.Flags)
	fmt.Printf("    - Broadcast:         %t\n", isBroadcast)
	fmt.Printf("    - Secondary Broadcast: %t\n", isSecondaryBroadcast)
	fmt.Printf("    - Merkle Tree Depth: %d\n", packet.MerkleTreeDepth)
	fmt.Printf("  - Epoch #:             %d\n", packet.Epoch)
	fmt.Printf("  - Timestamp (ms):      %d\n", packet.TimestampMs)
	fmt.Printf("  - App Message Hash:    %x\n", packet.AppMessageHash)
	fmt.Printf("  - App Message Length:  %d bytes\n", packet.AppMessageLen)

	// Merkle Proof
	fmt.Printf("  - Merkle Proof (%d items):\n", len(packet.MerkleProof))
	if len(packet.MerkleProof) > 0 {
		for i, hash := range packet.MerkleProof {
			fmt.Printf("    - Proof[%d]:            %x\n", i, hash)
		}
	} else {
		fmt.Printf("    - (None, Depth is %d)\n", packet.MerkleTreeDepth)
	}

	// 청크 특정 정보
	fmt.Printf("\n🧩 **Chunk Specific Data**\n")
	fmt.Printf("  - First Hop Recipient: %x\n", packet.FirstHopRecipient)
	fmt.Printf("  - Merkle Leaf Index:   %d\n", packet.MerkleLeafIdx)
	fmt.Printf("  - Reserved Byte:       0x%02x\n", packet.Reserved)
	fmt.Printf("  - Chunk ID:            %d\n", packet.ChunkID)
	fmt.Printf("  - Payload Length:      %d bytes\n", len(packet.Payload))
	if len(packet.Payload) > 0 {
		fmt.Printf("  - Payload (First 10B): %x...\n", packet.Payload[:min(10, len(packet.Payload))])
	}
	fmt.Println("----------------------------------")
}

func (data *Packet) NetworkHexDump() {
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf(" L2 (Ethernet) : %s -> %s (Type: %s)\n", data.EthernetLayer.SrcMAC, data.EthernetLayer.DstMAC, data.EthernetLayer.EthernetType)
	fmt.Printf(" L3 (IPv4)     : %s -> %s (Proto: %s)\n", data.IPv4Layer.SrcIP, data.IPv4Layer.DstIP, data.IPv4Layer.Protocol)

	if tcpLayer := data.TCPLayer; tcpLayer != nil {
		fmt.Printf(" L4 (TCP)      : Port %d -> %d\n", tcpLayer.SrcPort, tcpLayer.DstPort)
	} else if udpLayer := data.UDPLayer; udpLayer != nil {
		fmt.Printf(" L4 (UDP)      : Port %d -> %d\n", udpLayer.SrcPort, udpLayer.DstPort)
	} else {
		fmt.Printf(" L4 (Other)    : Protocol not TCP/UDP\n")
		return
	}
	fmt.Printf(" L7 (Payload)  : %d bytes\n", len(data.Payload))
	fmt.Println("-----------------------------------------------------------------")
}
