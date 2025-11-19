package consensus

import (
	"bytes"
	"fmt"
	"monad-flow/model/message/outbound_router/monad/consensus/protocol"

	"github.com/ethereum/go-ethereum/rlp"
)

type ConsensusMessage struct {
	Version uint32      `json:"version"`
	Payload interface{} `json:"payload,omitempty"`
}

func DecodeConsensusMessage(b []byte) (*ConsensusMessage, error) {
	s := rlp.NewStream(bytes.NewReader(b), uint64(len(b)))

	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("ConsensusMessage is not an RLP list: %w", err)
	}

	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("failed to open Inner List (Payload Container): %w", err)
	}

	version, err := s.Uint32()
	if err != nil {
		return nil, fmt.Errorf("failed to decode Version value: %w", err)
	}

	protoBytes, err := s.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to extract ProtocolMessage raw bytes: %w", err)
	}

	pMsg, err := protocol.DecodeProtocolMessage(protoBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ProtocolMessage: %w", err)
	}

	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("failed to close Inner List: %w", err)
	}

	msg := &ConsensusMessage{
		Version: version,
		Payload: pMsg,
	}

	return msg, nil
}
