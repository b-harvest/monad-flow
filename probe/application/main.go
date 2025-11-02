package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

//go:generate bpf2go bpf ../kernel/packet_capture.c -- -I/usr/include -I/usr/include/x86_64-linux-gnu

func main() {
	log.Println("eBPF 8000번 포트 필터링 프로그램을 시작합니다...")

	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	// 1. bpf 오브젝트 로드
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("eBPF 오브젝트 로딩 실패: %v", err)
	}

	// C 파일(packet_capture.c) 안의 함수가 'bpf_filter'라고 가정합니다.
	progFD := objs.BpfFilter.FD()
	defer objs.BpfFilter.Close()

	// 2. Raw 소켓 생성
	sock, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	if err != nil {
		log.Fatalf("Raw 소켓 생성 실패: %v", err)
	}
	defer syscall.Close(sock)

	// 3. Raw 소켓에 eBPF 필터 연결
	if err := syscall.SetsockoptInt(sock, syscall.SOL_SOCKET, syscall.SO_ATTACH_FILTER, progFD); err != nil {
		log.Fatalf("소켓에 eBPF 필터 연결 실패: %v", err)
	}
	log.Println("eBPF 필터가 소켓에 연결되었습니다. 포트 8000 캡처를 시작합니다...")

	// 4. 소켓에서 직접 패킷 읽기
	go func() {
		buf := make([]byte, 4096)
		for {
			n, _, err := syscall.Recvfrom(sock, buf, 0)
			if err != nil {
				log.Printf("소켓 읽기 중단: %v", err)
				return
			}

			// 5. gopacket으로 Raw 데이터 파싱
			packetData := buf[:n]
			packet := gopacket.NewPacket(packetData, layers.LayerTypeEthernet, gopacket.Default)

			// 6. 파싱된 정보 출력
			printPacketInfo(packet)
		}
	}()

	// Ctrl+C 신호를 받을 때까지 대기
	<-stopper
	log.Println("종료 신호를 받았습니다. 프로그램을 정리합니다.")
}

// gopacket이 파싱한 정보를 출력하는 함수
func printPacketInfo(packet gopacket.Packet) {
	var ip4 *layers.IPv4
	var tcp *layers.TCP
	var udp *layers.UDP
	var proto string

	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip4, _ = ipLayer.(*layers.IPv4)
	} else {
		return
	}

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ = tcpLayer.(*layers.TCP)
		proto = "TCP"
	} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ = udpLayer.(*layers.UDP)
		proto = "UDP"
	} else {
		return
	}

	fmt.Printf("\n--- [%s] %s:%s -> %s:%s ---\n",
		proto,
		ip4.SrcIP,
		getSrcPort(tcp, udp),
		ip4.DstIP,
		getDstPort(tcp, udp),
	)

	if appLayer := packet.ApplicationLayer(); appLayer != nil {
		fmt.Printf("Payload (%d bytes):\n", len(appLayer.Payload()))
		fmt.Println(hex.Dump(appLayer.Payload()))
	} else {
		fmt.Println("Payload (0 bytes)")
	}
}

// --- 헬퍼 함수 ---

func getSrcPort(tcp *layers.TCP, udp *layers.UDP) string {
	if tcp != nil {
		return fmt.Sprintf("%d", tcp.SrcPort)
	}
	if udp != nil {
		return fmt.Sprintf("%d", udp.SrcPort)
	}
	return "?"
}

func getDstPort(tcp *layers.TCP, udp *layers.UDP) string {
	if tcp != nil {
		return fmt.Sprintf("%d", tcp.DstPort)
	}
	if udp != nil {
		return fmt.Sprintf("%d", udp.DstPort)
	}
	return "?"
}

func htons(i uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return *(*uint16)(unsafe.Pointer(&b[0]))
}
