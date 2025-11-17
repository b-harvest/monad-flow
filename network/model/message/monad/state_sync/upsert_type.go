package state_sync

import "github.com/ethereum/go-ethereum/rlp"

type StateSyncUpsertType uint8

const (
	UpsertTypeCode          StateSyncUpsertType = 1
	UpsertTypeAccount       StateSyncUpsertType = 2
	UpsertTypeStorage       StateSyncUpsertType = 3
	UpsertTypeAccountDelete StateSyncUpsertType = 4
	UpsertTypeStorageDelete StateSyncUpsertType = 5
	UpsertTypeHeader        StateSyncUpsertType = 6
)

func (t *StateSyncUpsertType) DecodeRLP(s *rlp.Stream) error {
	if _, err := s.List(); err != nil {
		return err
	}
	var typeID uint8
	if err := s.Decode(&typeID); err != nil {
		return err
	}
	*t = StateSyncUpsertType(typeID)
	return s.ListEnd()
}
