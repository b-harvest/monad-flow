package monad

import (
	"fmt"
	"monad-flow/model/message/protocol"
	"monad-flow/model/message/protocol/advanced_round"
	"monad-flow/model/message/protocol/no_endorsement"
	"monad-flow/model/message/protocol/proposal"
	"monad-flow/model/message/protocol/round_recovery"
	"monad-flow/model/message/protocol/timeout"
	"monad-flow/model/message/protocol/vote"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type ConsensusMessage struct {
	Version uint32
	Message protocol.ProtocolMessage
}

func (cm *ConsensusMessage) DecodeRLP(s *rlp.Stream) error {
	var err error

	// 1. ConsensusMessage 리스트 디코딩: [Version, Message]
	listSize, err := s.List()
	if err != nil {
		return err
	}
	defer func() {
		if listSize > 0 {
			if endErr := s.ListEnd(); endErr != nil && err == nil {
				err = fmt.Errorf("failed to close ConsensusMessage list: %w", endErr)
			}
		}
	}()

	// 2. Version 디코딩 (Version 필드가 RLP List로 인코딩됨: [Version Value])
	versionListSize, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to decode Version list start: %w", err)
	}
	defer func() {
		if versionListSize > 0 {
			if endErr := s.ListEnd(); endErr != nil && err == nil {
				err = fmt.Errorf("failed to close Version list: %w", endErr)
			}
		}
	}()

	version, err := s.Uint32()
	if err != nil {
		return fmt.Errorf("failed to decode Version value (inside list): %w", err)
	}
	cm.Version = version

	// 3. Message (ProtocolMessage) 커스텀 디코딩: [Name, TypeID, Payload]
	messageListSize, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to decode ProtocolMessage list: %w", err)
	}
	defer func() {
		if messageListSize > 0 {
			if endErr := s.ListEnd(); endErr != nil && err == nil {
				err = fmt.Errorf("failed to close ProtocolMessage list: %w", endErr)
			}
		}
	}()

	nameBytes, err := s.Bytes()
	if err != nil {
		return fmt.Errorf("failed to decode ProtocolMessage name: %w", err)
	}

	name := string(nameBytes)
	if name != util.ProtocolMessageName {
		return fmt.Errorf("invalid protocol message name: %s", name)
	}

	typeID, err := s.Uint8()
	if err != nil {
		return fmt.Errorf("failed to decode ProtocolMessage TypeID: %w", err)
	}

	payloadBytes, err := s.Raw()
	if err != nil {
		return fmt.Errorf("failed to decode ProtocolMessage Payload: %w", err)
	}

	var msg protocol.ProtocolMessage
	switch typeID {
	case util.ProposalMsgType:
		msg = new(proposal.ProposalMessage)
	case util.VoteMsgType:
		msg = new(vote.VoteMessage)
	case util.TimeoutMsgType:
		msg = new(timeout.TimeoutMessage)
	case util.RoundRecoveryMsgType:
		msg = new(round_recovery.RoundRecoveryMessage)
	case util.NoEndorsementMsgType:
		msg = new(no_endorsement.NoEndorsementMessage)
	case util.AdvanceRoundMsgType:
		msg = new(advanced_round.AdvanceRoundMessage)
	default:
		return fmt.Errorf("unknown protocol message type ID: %d", typeID)
	}

	if err := rlp.DecodeBytes(payloadBytes, msg); err != nil {
		return fmt.Errorf("failed to decode message type %d payload: %w", typeID, err)
	}
	cm.Message = msg

	return nil
}
