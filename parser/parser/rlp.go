package parser

import (
	"bytes"
	"fmt"
	"log"
	"monad-flow/model/message/monad"
	"monad-flow/model/message/outbound_router"
	"monad-flow/model/message/outbound_router/fullnode_group"
	"monad-flow/model/message/outbound_router/peer_discovery"
	"monad-flow/model/message/protocol/advanced_round"
	"monad-flow/model/message/protocol/common"
	"monad-flow/model/message/protocol/no_endorsement"
	"monad-flow/model/message/protocol/proposal"
	"monad-flow/model/message/protocol/round_recovery"
	"monad-flow/model/message/protocol/timeout"
	"monad-flow/model/message/protocol/vote"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

func HandleDecodedMessage(data []byte) error {
	var outboundRouterMsg outbound_router.OutboundRouterMessage

	if err := rlp.Decode(bytes.NewReader(data), &outboundRouterMsg); err != nil {
		return fmt.Errorf("stage 1 (container) decode failed: %w", err)
	}

	switch outboundRouterMsg.MessageType {
	case util.AppMsgType:
		var monadMsg monad.MonadMessage
		if err := rlp.DecodeBytes(outboundRouterMsg.Message, &monadMsg); err != nil {
			return fmt.Errorf("stage 2 (AppMessage) decode failed: %w", err)
		}

		switch monadMsg.TypeID {
		case util.ConsensusMsgType:
			return handleConsensusMessage(monadMsg.Payload)
		case util.BlockSyncRequestMsgType:
			log.Println("[RLP-PARSE] Handling BlockSyncRequest (Type 2)... (not implemented)")
		case util.BlockSyncResponseMsgType:
			log.Println("[RLP-PARSE] Handling BlockSyncResponse (Type 3)... (not implemented)")
		case util.ForwardedTxMsgType:
			log.Println("[RLP-PARSE] Handling ForwardedTx (Type 4)... (not implemented)")
		case util.AdvanceRoundMsgType:
			log.Println("[RLP-PARSE] Handling StateSyncMessage (Type 5)... (not implemented)")
		default:
			return fmt.Errorf("unknown MonadMessage TypeID: %d", monadMsg.TypeID)
		}

	case util.PeerDiscType:
		peerDiscMsg, err := peer_discovery.DecodePeerDiscoveryMessage(outboundRouterMsg.Message)
		if err != nil {
			return fmt.Errorf("stage 2 (PeerDiscoveryMessage) decode failed: %w", err)
		}
		return handlePeerDiscoveryMessage(peerDiscMsg)

	case util.GroupType:
		fullNodeMsg, err := fullnode_group.DecodeFullNodesGroupMessage(outboundRouterMsg.Message)
		if err != nil {
			return fmt.Errorf("stage 2 (FullNodesGroup) decode failed: %w", err)
		}
		return handleFullNodesGroupMessage(fullNodeMsg)

	default:
		return fmt.Errorf("unknown OutboundRouter MessageType: %d", outboundRouterMsg.MessageType)
	}

	return nil
}

func handlePeerDiscoveryMessage(msg peer_discovery.PeerDiscoveryMessage) error {
	switch m := msg.(type) {
	case *peer_discovery.Ping:
		// log.Println("[RLP-PARSE] -> It's a Ping")
	case *peer_discovery.Pong:
		// log.Println("[RLP-PARSE] -> It's a Pong")
	case *peer_discovery.PeerLookupRequest:
		// log.Println("[RLP-PARSE] -> It's a PeerLookupRequest")
	case *peer_discovery.PeerLookupResponse:
		// log.Println("[RLP-PARSE] -> It's a PeerLookupResponse")
	case *peer_discovery.FullNodeRaptorcastRequest:
		log.Println("[RLP-PARSE] -> It's a FullNodeRaptorcastRequest")
	case *peer_discovery.FullNodeRaptorcastResponse:
		log.Println("[RLP-PARSE] -> It's a FullNodeRaptorcastResponse")
	default:
		return fmt.Errorf("unknown PeerDiscoveryMessage concrete type: %T", m)
	}

	return nil
}

func handleFullNodesGroupMessage(msg fullnode_group.FullNodesGroupMessage) error {
	switch m := msg.(type) {
	case *fullnode_group.PrepareGroup:
		// log.Println("[RLP-PARSE] -> It's a PrepareGroup")
	case *fullnode_group.PrepareGroupResponse:
		// log.Println("[RLP-PARSE] -> It's a PrepareGroupResponse")
	case *fullnode_group.ConfirmGroup:
		// log.Println("[RLP-PARSE] -> It's a ConfirmGroup")
	default:
		return fmt.Errorf("unknown FullNodesGroupMessage concrete type: %T", m)
	}

	return nil
}

func handleConsensusMessage(payload []byte) error {
	var consensusMsg monad.ConsensusMessage
	if err := rlp.Decode(bytes.NewReader(payload), &consensusMsg); err != nil {
		return fmt.Errorf("stage 2 (ConsensusMessage) decode failed: %w", err)
	}

	switch msg := consensusMsg.Message.(type) {
	case *proposal.ProposalMessage:
		// log.Printf("[RLP-PARSE]   -> IT'S A PROPOSAL! Round: %d, Epoch: %d", msg.ProposalRound, msg.ProposalEpoch)
	case *vote.VoteMessage:
		// log.Printf("[RLP-PARSE]   -> IT'S A VOTE! Round: %d, BlockID: %s", msg.Vote.Round, msg.Vote.ID.String())
	case *timeout.TimeoutMessage:
		log.Printf("[RLP-PARSE]   -> IT'S A TIMEOUT! Round: %d, Epoch: %d", msg.TMInfo.Round, msg.TMInfo.Epoch)
	case *round_recovery.RoundRecoveryMessage:
		log.Printf("[RLP-PARSE]   -> IT'S A ROUND RECOVERY! Round: %d, Epoch: %d", msg.Round, msg.Epoch)
	case *no_endorsement.NoEndorsementMessage:
		log.Printf("[RLP-PARSE]   -> IT'S A NO ENDORSEMENT! Round: %d, Epoch: %d", msg.Msg.Round, msg.Msg.Epoch)
	case *advanced_round.AdvanceRoundMessage:
		handleAdvancedRound(msg)
	default:
		log.Printf("[RLP-PARSE]   -> Unknown consensus message type: %T", msg)
	}

	return nil
}

func handleAdvancedRound(msg *advanced_round.AdvanceRoundMessage) error {
	switch arm := msg.LastRoundCertificate.Certificate.(type) {
	case *common.RoundCertificateQC:
		// log.Printf("[RLP-PARSE]   -> IT'S AN ADVANCE ROUND! QC round: %d", arm.QC.Info.Round)
	case *common.RoundCertificateTC:
		log.Printf("[RLP-PARSE]   -> IT'S AN ADVANCE ROUND! TC round: %d", arm.TC.Round)
	}
	return nil
}
