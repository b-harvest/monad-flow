package common

import (
	"fmt"
	"monad-flow/util"
	"net"

	"github.com/ethereum/go-ethereum/rlp"
)

// NameRecord는 Rust의 NameRecord 구조체에 해당합니다.
// SocketAddrV4를 IP와 Port로 분리했습니다.
type NameRecord struct {
	Address net.IP
	Port    uint16
	Seq     uint64
}

// MonadNameRecord는 Rust의 MonadNameRecord<ST> 구조체에 해당합니다.
// 이 구조체는 표준 RLP 인코딩(필드 리스트)을 사용합니다.
type MonadNameRecord struct {
	NameRecord *NameRecord
	Signature  util.Signature
}

func (nr *NameRecord) DecodeRLP(s *rlp.Stream) error {
    // 1. 전체 리스트 진입 ([IP, Port, Seq])
    _, err := s.List()
    if err != nil {
        return fmt.Errorf("NameRecord RLP is not a list: %w", err)
    }

    // 2. IP 디코딩 (Address)
    var ipBytes []byte
    if err = s.Decode(&ipBytes); err != nil {
        return fmt.Errorf("failed to decode NameRecord IP: %w", err)
    }
    if len(ipBytes) != 4 {
        return fmt.Errorf("decoded IP is not 4 bytes: got %d", len(ipBytes))
    }
    nr.Address = net.IP(ipBytes)

    // 3. Port 디코딩
    kind, _, err := s.Kind()
    if err != nil {
        return fmt.Errorf("failed to check kind for Port: %w", err)
    }

    if kind == rlp.List {
        if _, err = s.List(); err != nil {
            return fmt.Errorf("failed to open Port list: %w", err)
        }
        if err = s.Decode(&nr.Port); err != nil {
            return fmt.Errorf("failed to decode Port inside list: %w", err)
        }
        if err = s.ListEnd(); err != nil {
            return fmt.Errorf("failed to close Port list: %w", err)
        }
    } else {
        if err = s.Decode(&nr.Port); err != nil {
            return fmt.Errorf("failed to decode NameRecord Port: %w", err)
        }
    }

    // 4. Seq 디코딩
    if err = s.Decode(&nr.Seq); err != nil {
        return fmt.Errorf("failed to decode NameRecord Seq: %w", err)
    }

    // 5. 전체 리스트 종료
    return s.ListEnd()
}