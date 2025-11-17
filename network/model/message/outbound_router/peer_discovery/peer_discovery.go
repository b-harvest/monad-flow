package peer_discovery

import (
	"bytes"
	"fmt"
	"monad-flow/model/message/outbound_router/common"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type PeerDiscoveryMessage interface {
	implementsPeerDiscoveryMessage()
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

func (p *Ping) implementsPeerDiscoveryMessage()                       {}
func (p *Pong) implementsPeerDiscoveryMessage()                       {}
func (r *PeerLookupRequest) implementsPeerDiscoveryMessage()          {}
func (r *PeerLookupResponse) implementsPeerDiscoveryMessage()         {}
func (r *FullNodeRaptorcastRequest) implementsPeerDiscoveryMessage()  {}
func (r *FullNodeRaptorcastResponse) implementsPeerDiscoveryMessage() {}

func DecodePeerDiscoveryMessage(b []byte) (PeerDiscoveryMessage, error) {
	s := rlp.NewStream(bytes.NewReader(b), uint64(len(b)))

	_, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("peer discovery message is not an RLP list: %w", err)
	}

	// 1. 버전 디코딩
	var version uint16
	if err := s.Decode(&version); err != nil {
		return nil, fmt.Errorf("failed to decode peer discovery version: %w", err)
	}
	if version != util.PeerDiscoveryVersion {
		return nil, fmt.Errorf("unexpected peer discovery version: got %d, want %d", version, util.PeerDiscoveryVersion)
	}

	// 2. 메시지 타입 디코딩
	var msgType uint8
	if err := s.Decode(&msgType); err != nil {
		return nil, fmt.Errorf("failed to decode peer discovery message type: %w", err)
	}

	// 3. 타입에 따라 페이로드 디코딩
	var msg PeerDiscoveryMessage
	switch msgType {
	case util.PingMsgType:
		var p Ping
		if err := s.Decode(&p); err != nil {
			return nil, fmt.Errorf("failed to decode Ping payload: %w", err)
		}
		msg = &p
	case util.PongMsgType:
		var p Pong
		if err := s.Decode(&p); err != nil {
			return nil, fmt.Errorf("failed to decode Pong payload: %w", err)
		}
		msg = &p
	case util.PeerLookupRequestMsgType:
		var r PeerLookupRequest
		if err := s.Decode(&r); err != nil {
			return nil, fmt.Errorf("failed to decode PeerLookupRequest payload: %w", err)
		}
		msg = &r
	case util.PeerLookupResponseMsgType:
		var r PeerLookupResponse
		if err := s.Decode(&r); err != nil {
			return nil, fmt.Errorf("failed to decode PeerLookupResponse payload: %w", err)
		}
		msg = &r
	case util.FullNodeRaptorcastReqMsgType:
		msg = &FullNodeRaptorcastRequest{}
	case util.FullNodeRaptorcastRespMsgType:
		msg = &FullNodeRaptorcastResponse{}
	default:
		return nil, fmt.Errorf("unknown peer discovery message type: %d", msgType)
	}

	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("extra data after peer discovery message (type %d): %w", msgType, err)
	}

	return msg, nil
}
