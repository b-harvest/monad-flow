package monad

import (
	"bytes"
	"fmt"
	"monad-flow/model/message/outbound_router/monad/block_sync_request"
	"monad-flow/model/message/outbound_router/monad/block_sync_response"
	"monad-flow/model/message/outbound_router/monad/consensus"
	"monad-flow/model/message/outbound_router/monad/forwarded_tx"
	"monad-flow/model/message/outbound_router/monad/state_sync"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type MonadVersion struct {
	ProtocolVersion    uint32 `json:"protocolVersion"`
	ClientVersionMajor uint16 `json:"clientVersionMajor"`
	ClientVersionMinor uint16 `json:"clientVersionMinor"`
	HashVersion        uint16 `json:"hashVersion"`
	SerializeVersion   uint16 `json:"serializeVersion"`
}

type MonadMessage struct {
	Version MonadVersion `json:"version"`
	TypeID  uint8        `json:"typeId"`
	Payload interface{}  `json:"payload,omitempty"`
}

// DecodeMonadMessage: RLP 스트림을 이용해 헤더와 페이로드를 한 번에 파싱
func DecodeMonadMessage(b []byte) (*MonadMessage, error) {
	s := rlp.NewStream(bytes.NewReader(b), uint64(len(b)))

	_, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("MonadMessage is not an RLP list: %w", err)
	}

	var version MonadVersion
	if err := s.Decode(&version); err != nil {
		return nil, err
	}

	var typeID uint8
	if err := s.Decode(&typeID); err != nil {
		return nil, err
	}

	payloadBytes, err := s.Raw()
	if err != nil {
		return nil, err
	}

	msg := &MonadMessage{
		Version: version,
		TypeID:  typeID,
		Payload: nil,
	}

	switch typeID {

	case util.ConsensusMsgType:
		pMsg, err := consensus.DecodeConsensusMessage(payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode ConsensusMessage: %w", err)
		}
		msg.Payload = pMsg

	case util.ForwardedTxMsgType:
		txMsg, err := forwarded_tx.DecodeForwardedTxMessage(payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode ForwardedTxMessage: %w", err)
		}
		msg.Payload = txMsg

	case util.StateSyncMsgType:
		m, err := state_sync.HandleStateSyncMessage(payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to handle StateSyncMessage: %w", err)
		}
		msg.Payload = m

	case util.BlockSyncRequestMsgType:
		m, err := block_sync_request.HandleBlockSyncRequest(payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to handle BlockSyncRequest: %w", err)
		}
		msg.Payload = m

	case util.BlockSyncResponseMsgType:
		m, err := block_sync_response.HandleBlockSyncResponse(payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to handle BlockSyncResponse: %w", err)
		}
		msg.Payload = m

	default:
		return nil, fmt.Errorf("unknown MonadMessage TypeID: %d", typeID)
	}

	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("extra data after MonadMessage: %w", err)
	}

	return msg, nil
}
