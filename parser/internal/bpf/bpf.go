package bpf

import (
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/vishvananda/netlink"
)

// BPFMonitor는 eBPF 프로그램, 맵, 링크 등 관련 리소스를 캡슐화합니다.
type BPFMonitor struct {
	collection    *ebpf.Collection
	iface         netlink.Link
	qdiscIngress  netlink.Qdisc
	filterIngress netlink.Filter
	qdiscEgress   netlink.Qdisc
	filterEgress  netlink.Filter
	RingBufReader *ringbuf.Reader
}

// NewBPFMonitor는 eBPF 프로그램을 로드하고 TC 훅에 연결합니다.
func NewBPFMonitor(ifName string) (*BPFMonitor, error) {
	monitor := &BPFMonitor{}
	var err error

	// 1. 인터페이스 찾기
	monitor.iface, err = netlink.LinkByName(ifName)
	if err != nil {
		return nil, fmt.Errorf("failed to find interface %s: %w", ifName, err)
	}

	// 2. eBPF 프로그램 로드
	spec, err := ebpf.LoadCollectionSpec("./internal/packet_capture.o")
	if err != nil {
		return nil, fmt.Errorf("failed to load eBPF spec: %w", err)
	}

	monitor.collection, err = ebpf.NewCollection(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to create eBPF collection: %w", err)
	}

	// 3. 프로그램 가져오기
	progIngress := monitor.collection.Programs["tc_ingress"]
	if progIngress == nil {
		monitor.Close()
		return nil, fmt.Errorf("eBPF program 'tc_ingress' not found")
	}
	progEgress := monitor.collection.Programs["tc_egress"]
	if progEgress == nil {
		monitor.Close()
		return nil, fmt.Errorf("eBPF program 'tc_egress' not found")
	}

	// 4. Ingress 연결
	monitor.qdiscIngress, monitor.filterIngress, err = attachTC(monitor.iface, progIngress, netlink.HANDLE_MIN_INGRESS)
	if err != nil {
		monitor.Close()
		return nil, fmt.Errorf("failed to attach TC ingress: %w", err)
	}

	// 5. Egress 연결
	monitor.qdiscEgress, monitor.filterEgress, err = attachTC(monitor.iface, progEgress, netlink.HANDLE_MIN_EGRESS)
	if err != nil {
		monitor.Close()
		return nil, fmt.Errorf("failed to attach TC egress: %w", err)
	}

	log.Printf("Attached TC ingress to %s", ifName)
	log.Printf("Attached TC egress to %s", ifName)

	// 6. 링 버퍼 리더 생성
	monitor.RingBufReader, err = ringbuf.NewReader(monitor.collection.Maps["events"])
	if err != nil {
		monitor.Close()
		return nil, fmt.Errorf("failed to create ringbuf reader: %w", err)
	}

	return monitor, nil
}

// Close는 모든 eBPF 리소스(필터, qdisc, 맵)를 정리합니다.
func (m *BPFMonitor) Close() error {
	log.Println("Detaching eBPF programs and cleaning up...")
	var firstErr error

	if m.RingBufReader != nil {
		if err := m.RingBufReader.Close(); err != nil {
			firstErr = err
		}
	}
	if m.filterEgress != nil {
		if err := netlink.FilterDel(m.filterEgress); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if m.qdiscEgress != nil {
		if err := netlink.QdiscDel(m.qdiscEgress); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if m.filterIngress != nil {
		if err := netlink.FilterDel(m.filterIngress); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if m.qdiscIngress != nil {
		if err := netlink.QdiscDel(m.qdiscIngress); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if m.collection != nil {
		m.collection.Close()
	}

	return firstErr
}

// attachTC는 이 패키지 내부에서만 사용되는 헬퍼 함수입니다 (소문자).
func attachTC(iface netlink.Link, prog *ebpf.Program, direction uint32) (netlink.Qdisc, netlink.Filter, error) {
	qdisc := &netlink.GenericQdisc{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: iface.Attrs().Index,
			Handle:    netlink.MakeHandle(0xffff, 0),
			Parent:    netlink.HANDLE_CLSACT,
		},
		QdiscType: "clsact",
	}

	if err := netlink.QdiscAdd(qdisc); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, nil, fmt.Errorf("failed to add clsact qdisc: %w", err)
		}
	}

	filter := &netlink.BpfFilter{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: iface.Attrs().Index,
			Parent:    direction,
			Handle:    1,
			Protocol:  syscall.ETH_P_ALL,
		},
		Fd:           prog.FD(),
		Name:         prog.String(),
		DirectAction: true,
	}

	if err := netlink.FilterAdd(filter); err != nil {
		return nil, nil, fmt.Errorf("failed to add bpf filter: %w", err)
	}

	return qdisc, filter, nil
}
