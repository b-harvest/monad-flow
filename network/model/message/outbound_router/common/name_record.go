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

// DecodeRLP는 Rust의 커스텀 Decodable 구현을 따릅니다.
func (nr *NameRecord) DecodeRLP(s *rlp.Stream) error {
	// NameRecord가 [ip, port, seq] 리스트라고 가정합니다.
	_, err := s.List()
	if err != nil {
		return fmt.Errorf("NameRecord RLP is not a list: %w", err)
	}

	// 1. IP ([u8; 4])
	var ipBytes []byte
	if err = s.Decode(&ipBytes); err != nil {
		return fmt.Errorf("failed to decode NameRecord IP: %w", err)
	}
	if len(ipBytes) != 4 {
		return fmt.Errorf("decoded IP is not 4 bytes: got %d", len(ipBytes))
	}
	nr.Address = net.IP(ipBytes)

	// 2. Port (u16)
	if err = s.Decode(&nr.Port); err != nil {
		return fmt.Errorf("failed to decode NameRecord Port: %w", err)
	}

	// 3. Seq (u64)
	if err = s.Decode(&nr.Seq); err != nil {
		return fmt.Errorf("failed to decode NameRecord Seq: %w", err)
	}

	// 4. 리스트 종료
	return s.ListEnd()
}
