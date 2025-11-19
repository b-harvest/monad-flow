package state_sync

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/rlp"
)

type StateSyncRequest struct {
	Version     StateSyncVersion
	Prefix      uint64
	PrefixBytes uint8
	Target      uint64
	From        uint64
	Until       uint64
	OldTarget   uint64
}

func (req *StateSyncRequest) DecodeRLP(s *rlp.Stream) error {
	if _, err := s.List(); err != nil {
		return err
	}

	if err := s.Decode(&req.Version); err != nil {
		return fmt.Errorf("failed to decode version: %w", err)
	}

	if req.Version.IsCompatible() {
		if err := s.Decode(&req.Prefix); err != nil {
			return err
		}
		if err := s.Decode(&req.PrefixBytes); err != nil {
			return err
		}
		if err := s.Decode(&req.Target); err != nil {
			return err
		}
		if err := s.Decode(&req.From); err != nil {
			return err
		}
		if err := s.Decode(&req.Until); err != nil {
			return err
		}
		if err := s.Decode(&req.OldTarget); err != nil {
			return err
		}
	} else {
		log.Printf("Warning: Incompatible StateSyncRequest version %v detected. Skipping fields.", req.Version)
		var (
			prefix      uint64
			prefixBytes uint8
			target      uint64
			from        uint64
			until       uint64
			oldTarget   uint64
		)
		_ = s.Decode(&prefix)
		_ = s.Decode(&prefixBytes)
		_ = s.Decode(&target)
		_ = s.Decode(&from)
		_ = s.Decode(&until)
		_ = s.Decode(&oldTarget)
	}

	return s.ListEnd()
}
