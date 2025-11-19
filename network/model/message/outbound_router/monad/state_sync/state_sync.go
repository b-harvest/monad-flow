package state_sync

import (
	"fmt"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type SessionId struct {
	Value uint64
}

type stateSyncNetworkMessage struct {
	MessageName string
	TypeID      uint8
	Data        rlp.RawValue
}

type StateSyncNetworkMessage struct {
	IsRequest bool
	Request   StateSyncRequest

	IsResponse bool
	Response   StateSyncResponse

	IsBadVersion bool
	BadVersion   StateSyncBadVersion

	IsCompletion bool
	Completion   SessionId
}

func HandleStateSyncMessage(payload rlp.RawValue) (*StateSyncNetworkMessage, error) {
	var rlpMsg stateSyncNetworkMessage
	if err := rlp.DecodeBytes(payload, &rlpMsg); err != nil {
		return nil, fmt.Errorf("L5 (StateSync): failed to decode [Name, TypeID, Data]: %w", err)
	}

	if rlpMsg.MessageName != util.StateSyncMsgName {
		return nil, fmt.Errorf("L5 (StateSync): invalid message name: %s", rlpMsg.MessageName)
	}

	msg := &StateSyncNetworkMessage{}

	switch rlpMsg.TypeID {
	case util.TypeRequest:
		msg.IsRequest = true
		if err := rlp.DecodeBytes(rlpMsg.Data, &msg.Request); err != nil {
			return nil, fmt.Errorf("L5 (StateSync): failed to decode Request (Type 1): %w", err)
		}
	case util.TypeResponse:
		msg.IsResponse = true
		if err := rlp.DecodeBytes(rlpMsg.Data, &msg.Response); err != nil {
			return nil, fmt.Errorf("L5 (StateSync): failed to decode Response (Type 2): %w", err)
		}
	case util.TypeBadVersion:
		msg.IsBadVersion = true
		if err := rlp.DecodeBytes(rlpMsg.Data, &msg.BadVersion); err != nil {
			return nil, fmt.Errorf("L5 (StateSync): failed to decode BadVersion (Type 3): %w", err)
		}
	case util.TypeCompletion:
		msg.IsCompletion = true
		if err := rlp.DecodeBytes(rlpMsg.Data, &msg.Completion); err != nil {
			return nil, fmt.Errorf("L5 (StateSync): failed to decode Completion (Type 4): %w", err)
		}
	default:
		return nil, fmt.Errorf("L5 (StateSync): unknown TypeID: %d", rlpMsg.TypeID)
	}

	return msg, nil
}
