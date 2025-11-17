package block_sync_response

import (
	"bytes"
	"fmt"
	monad_common "monad-flow/model/message/monad/common"
	protocol_common "monad-flow/model/message/protocol/common"
	"monad-flow/model/message/protocol/proposal"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type BlockSyncResponse struct {
	IsHeadersResponse bool
	HeadersData       *BlockSyncHeadersResponse

	IsPayloadResponse bool
	PayloadData       *BlockSyncBodyResponse
}

type BlockSyncHeadersResponse struct {
	IsFound      bool
	FoundRange   monad_common.BlockRange
	FoundHeaders []*protocol_common.ConsensusBlockHeader

	IsNotAvailable bool
	NotAvailRange  monad_common.BlockRange
}

type BlockSyncBodyResponse struct {
	IsFound   bool
	FoundBody *proposal.ConsensusBlockBody

	IsNotAvailable  bool
	NotAvailPayload util.ConsensusBlockBodyId
}

type blockSyncResponseMessage struct {
	MessageName string
	TypeID      uint8
	Data        rlp.RawValue
}

func HandleBlockSyncResponse(payload rlp.RawValue) (*BlockSyncResponse, error) {
	// 1. RLP 리스트를 디코딩합니다.
	var rlpResponse blockSyncResponseMessage
	if err := rlp.DecodeBytes(payload, &rlpResponse); err != nil {
		return nil, fmt.Errorf("L5 (BlockSyncResponse): failed to RLP-decode [Name, TypeID, Data]: %w", err)
	}

	// 2. 메시지 이름을 검증합니다.
	if rlpResponse.MessageName != util.BlockSyncResMsgName {
		return nil, fmt.Errorf("L5 (BlockSyncResponse): unexpected message name: %s", rlpResponse.MessageName)
	}

	// 3. 타입 ID에 따라 파서를 호출합니다.
	finalResponse := &BlockSyncResponse{}
	switch rlpResponse.TypeID {
	case util.BlockSyncHeaderType:
		finalResponse.IsHeadersResponse = true
		headersData, err := parseHeadersResponse(rlpResponse.Data)
		if err != nil {
			return nil, err
		}
		finalResponse.HeadersData = headersData

	case util.BlockSyncBodyType:
		finalResponse.IsPayloadResponse = true
		payloadData, err := parseBodyResponse(rlpResponse.Data)
		if err != nil {
			return nil, err
		}
		finalResponse.PayloadData = payloadData

	default:
		return nil, fmt.Errorf("L5 (BlockSyncResponse): unknown TypeID: %d", rlpResponse.TypeID)
	}

	return finalResponse, nil
}

func parseHeadersResponse(rlpData rlp.RawValue) (*BlockSyncHeadersResponse, error) {
	s := rlp.NewStream(bytes.NewReader(rlpData), 0)

	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("L6 (HeadersResponse): expected RLP list: %w", err)
	}

	var msgName string
	if err := s.Decode(&msgName); err != nil {
		return nil, fmt.Errorf("L6 (HeadersResponse): failed to decode MessageName: %w", err)
	}
	if msgName != util.BlockSyncHdrResName {
		return nil, fmt.Errorf("L6 (HeadersResponse): name mismatch: got %s", msgName)
	}

	var typeID uint8
	if err := s.Decode(&typeID); err != nil {
		return nil, fmt.Errorf("L6 (HeadersResponse): failed to decode TypeID: %w", err)
	}

	resp := &BlockSyncHeadersResponse{}
	switch typeID {
	case util.Found:
		resp.IsFound = true
		if err := s.Decode(&resp.FoundRange); err != nil {
			return nil, fmt.Errorf("L6 (HeadersResponse/Found): failed to decode Range: %w", err)
		}
		if err := s.Decode(&resp.FoundHeaders); err != nil {
			return nil, fmt.Errorf("L6 (HeadersResponse/Found): failed to decode Headers List: %w", err)
		}

	case util.NotAvailable:
		resp.IsNotAvailable = true
		if err := s.Decode(&resp.NotAvailRange); err != nil {
			return nil, fmt.Errorf("L6 (HeadersResponse/NotAvail): failed to decode Range: %w", err)
		}

	default:
		return nil, fmt.Errorf("L6 (HeadersResponse): unknown TypeID: %d", typeID)
	}

	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("L6 (HeadersResponse): RLP list has trailing data: %w", err)
	}

	return resp, nil
}

func parseBodyResponse(rlpData rlp.RawValue) (*BlockSyncBodyResponse, error) {
	s := rlp.NewStream(bytes.NewReader(rlpData), 0)

	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("L6 (BodyResponse): expected RLP list: %w", err)
	}

	var msgName string
	if err := s.Decode(&msgName); err != nil {
		return nil, fmt.Errorf("L6 (BodyResponse): failed to decode MessageName: %w", err)
	}
	if msgName != util.BlockSyncBdyResName {
		return nil, fmt.Errorf("L6 (BodyResponse): name mismatch: got %s", msgName)
	}

	var typeID uint8
	if err := s.Decode(&typeID); err != nil {
		return nil, fmt.Errorf("L6 (BodyResponse): failed to decode TypeID: %w", err)
	}

	resp := &BlockSyncBodyResponse{}
	switch typeID {
	case util.Found:
		resp.IsFound = true
		if err := s.Decode(&resp.FoundBody); err != nil {
			return nil, fmt.Errorf("L6 (BodyResponse/Found): failed to decode Body: %w", err)
		}

	case util.NotAvailable:
		resp.IsNotAvailable = true
		if err := s.Decode(&resp.NotAvailPayload); err != nil {
			return nil, fmt.Errorf("L6 (BodyResponse/NotAvail): failed to decode PayloadID: %w", err)
		}

	default:
		return nil, fmt.Errorf("L6 (BodyResponse): unknown TypeID: %d", typeID)
	}

	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("L6 (BodyResponse): RLP list has trailing data: %w", err)
	}

	return resp, nil
}
