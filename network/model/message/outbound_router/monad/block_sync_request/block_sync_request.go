package block_sync_request

import (
	"fmt"
	"monad-flow/model/message/outbound_router/monad/common"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type BlockSyncRequest struct {
	MessageName string
	TypeID      uint8

	IsHeaders bool
	Headers   common.BlockRange

	IsPayload bool
	Payload   util.ConsensusBlockBodyId
}

type blockSyncRequestMessage struct {
	MessageName string
	TypeID      uint8
	Data        rlp.RawValue
}

func HandleBlockSyncRequest(payload rlp.RawValue) (*BlockSyncRequest, error) {
	// 1. L5 RLP 리스트 [Name, TypeID, Data]를 디코딩합니다.
	var rlpRequest blockSyncRequestMessage
	if err := rlp.DecodeBytes(payload, &rlpRequest); err != nil {
		return nil, fmt.Errorf("L5 (BlockSync): failed to RLP-decode [Name, TypeID, Data]: %w", err)
	}

	// 2. 메시지 이름을 검증합니다.
	if rlpRequest.MessageName != util.BlockSyncReqMsgName {
		return nil, fmt.Errorf("L5 (BlockSync): unexpected message name: %s", rlpRequest.MessageName)
	}

	// 3. 타입 ID에 따라 최종 데이터를 파싱합니다.
	finalRequest := &BlockSyncRequest{}
	switch rlpRequest.TypeID {
	case util.BlockSyncHeaderType:
		finalRequest.IsHeaders = true
		if err := rlp.DecodeBytes(rlpRequest.Data, &finalRequest.Headers); err != nil {
			return nil, fmt.Errorf("L5 (BlockSync): failed to RLP-decode BlockRange: %w", err)
		}

	case util.BlockSyncBodyType:
		finalRequest.IsPayload = true
		if err := rlp.DecodeBytes(rlpRequest.Data, &finalRequest.Payload); err != nil {
			return nil, fmt.Errorf("L5 (BlockSync): failed to RLP-decode ConsensusBlockBodyId: %w", err)
		}

	default:
		return nil, fmt.Errorf("L5 (BlockSync): unknown TypeID: %d", rlpRequest.TypeID)
	}

	return finalRequest, nil
}
