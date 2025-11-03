package network

import (
	"fmt"
	"monad-flow/internal/monad"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func ParseAndDumpPacket(data []byte) {
	// eBPF가 L2(Ethernet)부터 캡처했으므로 LinkTypeEthernet 사용
	packet := gopacket.NewPacket(data, layers.LinkTypeEthernet, gopacket.Default)

	// L2 (Ethernet)
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth, _ := ethLayer.(*layers.Ethernet)
		fmt.Printf(" L2 (Ethernet) : %s -> %s (Type: %s)\n", eth.SrcMAC, eth.DstMAC, eth.EthernetType)
	}

	// L3 (IPv4)
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		fmt.Printf(" L3 (IPv4)     : %s -> %s (Proto: %s)\n", ip.SrcIP, ip.DstIP, ip.Protocol)
	}

	// L4 (TCP / UDP)
	var l4Proto string
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		fmt.Printf(" L4 (TCP)      : Port %d -> %d\n", tcp.SrcPort, tcp.DstPort)
		l4Proto = "TCP"
	} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		fmt.Printf(" L4 (UDP)      : Port %d -> %d\n", udp.SrcPort, udp.DstPort)
		l4Proto = "UDP"
	} else {
		fmt.Printf(" L4 (Other)    : Protocol not TCP/UDP\n")
		return
	}

	// L7 (Application Payload)
	if appLayer := packet.ApplicationLayer(); appLayer != nil {
		payload := appLayer.Payload()
		if len(payload) > 0 {
			fmt.Printf(" L7 (Payload)  : %s, %d bytes\n", l4Proto, len(payload))
			fmt.Println("-----------------------------------------------------------------")
			monad.HexDumpAndReassemble(payload)
			fmt.Println("-----------------------------------------------------------------")
		} else {
			fmt.Printf(" L7 (Payload)  : [No Payload]\n")
		}
	} else {
		fmt.Printf(" L7 (Payload)  : [No Application Layer]\n")
	}
}
