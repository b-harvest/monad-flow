package state_sync

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

// StateSyncUpsertV1 (표준 RLP)
type StateSyncUpsertV1 struct {
	UpsertType StateSyncUpsertType
	Data       []byte
}

// StateSyncUpsertV0 (V1 호환성용)
type StateSyncUpsertV0 struct {
	UpsertType StateSyncUpsertType
	Data       []byte
}

type StateSyncResponse struct {
	Version       StateSyncVersion
	Nonce         uint64
	ResponseIndex uint32
	Request       StateSyncRequest
	Response      []StateSyncUpsertV1
	ResponseN     uint64
}

func (resp *StateSyncResponse) DecodeRLP(s *rlp.Stream) error {
	if _, err := s.List(); err != nil {
		return err
	}

	if err := s.Decode(&resp.Version); err != nil {
		return err
	}
	if err := s.Decode(&resp.Nonce); err != nil {
		return err
	}
	if err := s.Decode(&resp.ResponseIndex); err != nil {
		return err
	}
	if err := s.Decode(&resp.Request); err != nil {
		return fmt.Errorf("failed to decode nested Request: %w", err)
	}

	if resp.Version.Ge(STATESYNC_VERSION_V1) {
		if err := s.Decode(&resp.Response); err != nil {
			return fmt.Errorf("failed to decode V1 upsert list: %w", err)
		}
	} else {
		var v0Response []StateSyncUpsertV0
		if err := s.Decode(&v0Response); err != nil {
			return fmt.Errorf("failed to decode V0 upsert list: %w", err)
		}
		resp.Response = make([]StateSyncUpsertV1, len(v0Response))
		for i, v0 := range v0Response {
			resp.Response[i] = StateSyncUpsertV1{
				UpsertType: v0.UpsertType,
				Data:       v0.Data,
			}
		}
	}

	if err := s.Decode(&resp.ResponseN); err != nil {
		return err
	}
	return s.ListEnd()
}
