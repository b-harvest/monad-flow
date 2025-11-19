package peer_discovery

import (
	"bytes"
	"fmt"
	"monad-flow/model/message/outbound_router/common"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type PeerDiscoveryMessage struct {
	Version uint16      `json:"version"`
	Type    uint8       `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type Ping struct {
	ID              uint32
	LocalNameRecord *common.MonadNameRecord
}

type Pong struct {
	PingID         uint32
	LocalRecordSeq uint64
}

type PeerLookupRequest struct {
	LookupID      uint32
	Target        util.NodeID
	OpenDiscovery bool
}

type PeerLookupResponse struct {
	LookupID    uint32
	Target      util.NodeID
	NameRecords []*common.MonadNameRecord
}

type FullNodeRaptorcastRequest struct{}

type FullNodeRaptorcastResponse struct{}

func DecodePeerDiscoveryMessage(b []byte) (*PeerDiscoveryMessage, error) {
	s := rlp.NewStream(bytes.NewReader(b), uint64(len(b)))

	// Start list
	_, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("PeerDiscovery RLP is not a list: %w", err)
	}

	// Version
	var version uint16
	if err := s.Decode(&version); err != nil {
		return nil, fmt.Errorf("failed to decode PeerDiscovery version: %w", err)
	}
	if version != util.PeerDiscoveryVersion {
		return nil, fmt.Errorf("unexpected PeerDiscovery version: got %d want %d",
			version, util.PeerDiscoveryVersion)
	}

	// Type
	var msgType uint8
	if err := s.Decode(&msgType); err != nil {
		return nil, fmt.Errorf("failed to decode PeerDiscovery type: %w", err)
	}

	// Extract raw payload
	payloadBytes, err := s.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to extract PeerDiscovery payload: %w", err)
	}

	msg := &PeerDiscoveryMessage{
		Version: version,
		Type:    msgType,
		Payload: nil,
	}

	// Decode according to type
	switch msgType {

	case util.PingMsgType:
		var p Ping
		if err := rlp.DecodeBytes(payloadBytes, &p); err != nil {
			return nil, fmt.Errorf("failed to decode Ping: %w", err)
		}
		msg.Payload = &p

	case util.PongMsgType:
		var p Pong
		if err := rlp.DecodeBytes(payloadBytes, &p); err != nil {
			return nil, fmt.Errorf("failed to decode Pong: %w", err)
		}
		msg.Payload = &p

	case util.PeerLookupRequestMsgType:
		var req PeerLookupRequest
		if err := rlp.DecodeBytes(payloadBytes, &req); err != nil {
			return nil, fmt.Errorf("failed to decode PeerLookupRequest: %w", err)
		}
		msg.Payload = &req

	case util.PeerLookupResponseMsgType:
		var resp PeerLookupResponse
		if err := rlp.DecodeBytes(payloadBytes, &resp); err != nil {
			return nil, fmt.Errorf("failed to decode PeerLookupResponse: %w", err)
		}
		msg.Payload = &resp

	case util.FullNodeRaptorcastReqMsgType:
		msg.Payload = &FullNodeRaptorcastRequest{}

	case util.FullNodeRaptorcastRespMsgType:
		msg.Payload = &FullNodeRaptorcastResponse{}

	default:
		return nil, fmt.Errorf("unknown PeerDiscovery type: %d", msgType)
	}

	// End list
	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("extra data after PeerDiscovery message: %w", err)
	}

	return msg, nil
}
