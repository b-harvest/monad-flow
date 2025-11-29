package common

import (
	"fmt"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type WirePort struct {
    Tag  uint8
    Port uint16
}

type WireNameRecordV1 struct {
    IP   []byte
    Port uint16
    Seq  uint64
}

type WireNameRecordV2 struct {
    IP           []byte
    Ports        []WirePort
    Capabilities uint64
    Seq          uint64
}

type VersionedNameRecord interface {}

type NameRecord struct {
    Record VersionedNameRecord
}

type MonadNameRecord struct {
    NameRecord *NameRecord
    Signature  util.Signature
}

func (nr *NameRecord) DecodeRLP(s *rlp.Stream) error {
    // 1. 원본 바이트 읽기 
    raw, err := s.Raw()
    if err != nil {
        return fmt.Errorf("failed to read raw bytes: %w", err)
    }

    // 2. V2 디코딩 시도 
    var v2 WireNameRecordV2
    if err := rlp.DecodeBytes(raw, &v2); err == nil {
        if len(v2.IP) != 4 {
             return fmt.Errorf("invalid V2 IPv4 length")
        }
        nr.Record = &v2
        return nil
    }

    // 3. 실패 시 V1 디코딩 시도
    var v1 WireNameRecordV1
    if err := rlp.DecodeBytes(raw, &v1); err == nil {
        if len(v1.IP) != 4 {
             return fmt.Errorf("invalid V1 IPv4 length")
        }
        nr.Record = &v1
        return nil
    }

    // 4. 둘 다 실패
    return fmt.Errorf("failed to decode NameRecord (neither V1 nor V2)")
}