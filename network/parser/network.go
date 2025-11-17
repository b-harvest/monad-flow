package parser

import (
	"monad-flow/model"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func ParsePacket(data []byte) model.Packet {
	info := model.Packet{}
	packet := gopacket.NewPacket(data, layers.LinkTypeEthernet, gopacket.Default)

	// L2 (Ethernet)
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		info.EthernetLayer = ethLayer.(*layers.Ethernet)
	}

	// L3 (IPv4)
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		info.IPv4Layer = ipLayer.(*layers.IPv4)
	}

	// L4 (TCP / UDP)
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		info.TCPLayer = tcpLayer.(*layers.TCP)
	} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		info.UDPLayer = udpLayer.(*layers.UDP)
	}

	// L7 (Application Payload)
	if appLayer := packet.ApplicationLayer(); appLayer != nil {
		info.Payload = appLayer.Payload()
	}

	return info
}
