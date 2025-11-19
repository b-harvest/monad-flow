package protocol

import (
	"bytes"
	"fmt"
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/advanced_round"
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/no_endorsement"
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/proposal"
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/round_recovery"
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/timeout"
	"monad-flow/model/message/outbound_router/monad/consensus/protocol/vote"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type ProtocolMessage struct {
	Name        string      `json:"name"`
	MessageType uint8       `json:"messageType"`
	Payload     interface{} `json:"payload"`
}

func DecodeProtocolMessage(b []byte) (*ProtocolMessage, error) {
	s := rlp.NewStream(bytes.NewReader(b), uint64(len(b)))

	// 1. ProtocolMessage 리스트 시작 [Name, Type, Payload]
	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("ProtocolMessage is not an RLP list: %w", err)
	}

	// 2. Name 디코딩 (String)
	var name string
	if err := s.Decode(&name); err != nil {
		return nil, fmt.Errorf("failed to decode ProtocolMessage Name: %w", err)
	}

	if name != util.ProtocolMessageName {
		return nil, fmt.Errorf("invalid protocol message name: '%s'", name)
	}

	// 3. MessageType 디코딩 (Uint8)
	var msgType uint8
	if err := s.Decode(&msgType); err != nil {
		return nil, fmt.Errorf("failed to decode ProtocolMessage MessageType: %w", err)
	}

	// 4. Payload Raw 바이트 추출
	payloadBytes, err := s.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to extract ProtocolMessage Payload: %w", err)
	}

	msg := &ProtocolMessage{
		Name:        name,
		MessageType: msgType,
		Payload:     nil,
	}

	switch msgType {
	case util.ProposalMsgType:
		var p proposal.ProposalMessage
		if err := rlp.DecodeBytes(payloadBytes, &p); err != nil {
			return nil, fmt.Errorf("failed to decode ProposalMessage: %w", err)
		}
		msg.Payload = &p

	case util.VoteMsgType:
		var v vote.VoteMessage
		if err := rlp.DecodeBytes(payloadBytes, &v); err != nil {
			return nil, fmt.Errorf("failed to decode VoteMessage: %w", err)
		}
		msg.Payload = &v

	case util.TimeoutMsgType:
		var t timeout.TimeoutMessage
		if err := rlp.DecodeBytes(payloadBytes, &t); err != nil {
			return nil, fmt.Errorf("failed to decode TimeoutMessage: %w", err)
		}
		msg.Payload = &t

	case util.RoundRecoveryMsgType:
		var rr round_recovery.RoundRecoveryMessage
		if err := rlp.DecodeBytes(payloadBytes, &rr); err != nil {
			return nil, fmt.Errorf("failed to decode RoundRecoveryMessage: %w", err)
		}
		msg.Payload = &rr

	case util.NoEndorsementMsgType:
		var ne no_endorsement.NoEndorsementMessage
		if err := rlp.DecodeBytes(payloadBytes, &ne); err != nil {
			return nil, fmt.Errorf("failed to decode NoEndorsementMessage: %w", err)
		}
		msg.Payload = &ne

	case util.AdvanceRoundMsgType:
		var ar advanced_round.AdvanceRoundMessage
		if err := rlp.DecodeBytes(payloadBytes, &ar); err != nil {
			return nil, fmt.Errorf("failed to decode AdvanceRoundMessage: %w", err)
		}
		msg.Payload = &ar

	default:
		return nil, fmt.Errorf("unknown protocol message type ID: %d", msgType)
	}

	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("extra data after ProtocolMessage: %w", err)
	}

	return msg, nil
}
