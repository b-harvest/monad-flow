package fullnode_group

import (
	"bytes"
	"fmt"
	"monad-flow/model/message/outbound_router/common"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type FullNodesGroupMessage struct {
	Version uint8       `json:"version"`
	Type    uint8       `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type PrepareGroup struct {
	ValidatorID  util.NodeID
	MaxGroupSize uint64
	StartRound   util.Round
	EndRound     util.Round
}

type PrepareGroupResponse struct {
	Req    *PrepareGroup
	NodeID util.NodeID
	Accept bool
}

type ConfirmGroup struct {
	Prepare     *PrepareGroup
	Peers       []util.NodeID
	NameRecords []*common.MonadNameRecord
}

func DecodeFullNodesGroupMessage(b []byte) (*FullNodesGroupMessage, error) {
	s := rlp.NewStream(bytes.NewReader(b), uint64(len(b)))

	// 1. 리스트 시작 [version, type, payload]
	_, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("FullNodesGroup message is not an RLP list: %w", err)
	}

	// 2. 버전 디코딩
	var version uint8
	if err := s.Decode(&version); err != nil {
		return nil, fmt.Errorf("failed to decode group message version: %w", err)
	}
	if version != util.GroupMsgVersion {
		return nil, fmt.Errorf("unknown group message version: got %d, want %d", version, util.GroupMsgVersion)
	}

	// 3. 메시지 타입 디코딩
	var msgType uint8
	if err := s.Decode(&msgType); err != nil {
		return nil, fmt.Errorf("failed to decode group message type: %w", err)
	}

	// 4. Payload 부분의 Raw 바이트 추출
	payloadBytes, err := s.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to extract group message payload: %w", err)
	}

	// 반환할 메시지 객체 생성
	msg := &FullNodesGroupMessage{
		Version: version,
		Type:    msgType,
		Payload: nil,
	}

	// 5. 타입에 따라 페이로드 디코딩
	switch msgType {
	case util.MsgTypePrepReq:
		var p PrepareGroup
		if err := rlp.DecodeBytes(payloadBytes, &p); err != nil {
			return nil, fmt.Errorf("failed to decode PrepareGroup payload: %w", err)
		}
		msg.Payload = &p

	case util.MsgTypePrepRes:
		var r PrepareGroupResponse
		if err := rlp.DecodeBytes(payloadBytes, &r); err != nil {
			return nil, fmt.Errorf("failed to decode PrepareGroupResponse payload: %w", err)
		}
		msg.Payload = &r

	case util.MsgTypeConfGrp:
		var c ConfirmGroup
		if err := rlp.DecodeBytes(payloadBytes, &c); err != nil {
			return nil, fmt.Errorf("failed to decode ConfirmGroup payload: %w", err)
		}
		msg.Payload = &c

	default:
		return nil, fmt.Errorf("unknown FullNodesGroup message type: %d", msgType)
	}

	// 6. 리스트 끝 확인
	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("extra data after group message payload (type %d): %w", msgType, err)
	}

	return msg, nil
}
