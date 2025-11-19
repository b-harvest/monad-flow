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
	MessageName string
	TypeID      uint8

	Request    StateSyncRequest
	Response   StateSyncResponse
	BadVersion StateSyncBadVersion
	Completion SessionId
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
	msg.MessageName = rlpMsg.MessageName
	msg.TypeID = rlpMsg.TypeID

	switch rlpMsg.TypeID {
	case util.TypeRequest:
		if err := rlp.DecodeBytes(rlpMsg.Data, &msg.Request); err != nil {
			return nil, fmt.Errorf("L5 (StateSync): failed to decode Request (Type 1): %w", err)
		}
	case util.TypeResponse:
		if err := rlp.DecodeBytes(rlpMsg.Data, &msg.Response); err != nil {
			return nil, fmt.Errorf("L5 (StateSync): failed to decode Response (Type 2): %w", err)
		}
	case util.TypeBadVersion:
		if err := rlp.DecodeBytes(rlpMsg.Data, &msg.BadVersion); err != nil {
			return nil, fmt.Errorf("L5 (StateSync): failed to decode BadVersion (Type 3): %w", err)
		}
	case util.TypeCompletion:
		if err := rlp.DecodeBytes(rlpMsg.Data, &msg.Completion); err != nil {
			return nil, fmt.Errorf("L5 (StateSync): failed to decode Completion (Type 4): %w", err)
		}
	default:
		return nil, fmt.Errorf("L5 (StateSync): unknown TypeID: %d", rlpMsg.TypeID)
	}

	return msg, nil
}
