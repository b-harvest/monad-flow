package fullnode_group

import (
	"bytes"
	"fmt"
	"monad-flow/model/message/outbound_router/common"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type FullNodesGroupMessage interface {
	implementsFullNodesGroupMessage()
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

func (m *ConfirmGroup) implementsFullNodesGroupMessage()         {}
func (m *PrepareGroup) implementsFullNodesGroupMessage()         {}
func (m *PrepareGroupResponse) implementsFullNodesGroupMessage() {}

func DecodeFullNodesGroupMessage(b []byte) (FullNodesGroupMessage, error) {
	s := rlp.NewStream(bytes.NewReader(b), uint64(len(b)))

	// 1. 래퍼 리스트 디코딩 [version, type, payload]
	_, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("FullNodesGroup message is not an RLP list: %w", err)
	}

	// 2. 버전 디코딩 및 확인
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

	// 4. 타입에 따라 페이로드 디코딩
	var msg FullNodesGroupMessage
	switch msgType {
	case util.MsgTypePrepReq:
		var p PrepareGroup
		if err := s.Decode(&p); err != nil {
			return nil, fmt.Errorf("failed to decode PrepareGroup payload: %w", err)
		}
		msg = &p
	case util.MsgTypePrepRes:
		var r PrepareGroupResponse
		if err := s.Decode(&r); err != nil {
			return nil, fmt.Errorf("failed to decode PrepareGroupResponse payload: %w", err)
		}
		msg = &r
	case util.MsgTypeConfGrp:
		var c ConfirmGroup
		if err := s.Decode(&c); err != nil {
			return nil, fmt.Errorf("failed to decode ConfirmGroup payload: %w", err)
		}
		msg = &c
	default:
		return nil, fmt.Errorf("unknown FullNodesGroup message type: %d", msgType)
	}

	// 5. 래퍼 리스트 끝 확인
	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("extra data after group message payload (type %d): %w", msgType, err)
	}

	return msg, nil
}
