package util

import (
	"bytes"
	"fmt"

	"monad-flow/model"
)

func NetworkHexDump(data model.Packet) {
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

func PrintMonadPacketDetails(packet *model.MonadChunkPacket) {
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

func ApplicationHexDump(data []byte) {
	fmt.Printf(" L7 (Payload)  : %d bytes\n", len(data))
	fmt.Println("-----------------------------------------------------------------")
	const bytesPerLine = 16
	var hexBuf, asciiBuf bytes.Buffer

	for i := 0; i < len(data); i += bytesPerLine {
		fmt.Printf("  %08x: ", i)

		hexBuf.Reset()
		asciiBuf.Reset()

		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}

		line := data[i:end]

		for j := 0; j < len(line); j++ {
			hexBuf.WriteString(fmt.Sprintf("%02x ", line[j]))
			if j == 7 {
				hexBuf.WriteString(" ")
			}
		}

		for _, b := range line {
			if b >= 32 && b <= 126 {
				asciiBuf.WriteByte(b)
			} else {
				asciiBuf.WriteByte('.')
			}
		}

		hexStr := hexBuf.String()
		padding := (bytesPerLine * 3) + (bytesPerLine / 8)
		fmt.Printf("%-*s |%s|\n", padding, hexStr, asciiBuf.String())
	}
	fmt.Println("-----------------------------------------------------------------")
}
