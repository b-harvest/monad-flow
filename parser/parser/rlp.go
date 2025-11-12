package parser

import (
	"bytes"
	"fmt"
	"log"
	"monad-flow/model/message/monad"
	"monad-flow/model/message/monad/block_sync_request"
	"monad-flow/model/message/monad/block_sync_response"
	"monad-flow/model/message/monad/consensus"
	"monad-flow/model/message/monad/forwarded_tx"
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
			req, err := block_sync_request.HandleBlockSyncRequest(monadMsg.Payload)
			if err != nil {
				return err
			}
			if req.IsHeaders {
				log.Printf("    L5 Type: Headers Request - LastBlockID: %s,  NumBlocks: %d", req.Headers.LastBlockID.Hex(), req.Headers.NumBlocks)
			} else if req.IsPayload {
				log.Printf("    L5 Type: Payload Request - PayloadID: %s", req.Payload.Hex())
			}
		case util.BlockSyncResponseMsgType:
			resp, err := block_sync_response.HandleBlockSyncResponse(monadMsg.Payload)
			if err != nil {
				return err
			}
			if resp.IsHeadersResponse {
				if resp.HeadersData.IsFound {
					log.Printf("    L5/L6 Type: HeadersResponse (Found)")
					log.Printf("      Range: %d blocks, last ID %s", resp.HeadersData.FoundRange.NumBlocks, resp.HeadersData.FoundRange.LastBlockID.Hex())
					log.Printf("      Headers Rcvd: %d", len(resp.HeadersData.FoundHeaders))
				} else if resp.HeadersData.IsNotAvailable {
					log.Printf("    L5/L6 Type: HeadersResponse (NotAvailable)")
					log.Printf("      Range: %d blocks, last ID %s", resp.HeadersData.NotAvailRange.NumBlocks, resp.HeadersData.NotAvailRange.LastBlockID.Hex())
				}
			} else if resp.IsPayloadResponse {
				if resp.PayloadData.IsFound {
					log.Printf("    L5/L6 Type: PayloadResponse (Found)")
					log.Printf("      Transaction len: %d", len(resp.PayloadData.FoundBody.ExecutionBody.Transactions))
				} else if resp.PayloadData.IsNotAvailable {
					log.Printf("    L5/L6 Type: PayloadResponse (NotAvailable)")
					log.Printf("      PayloadID: %s", resp.PayloadData.NotAvailPayload.Hex())
				}
			}
		case util.ForwardedTxMsgType:
			return handleForwardedTx(monadMsg.Payload)
		case util.StateSyncMsgType:
			// log.Println("[RLP-PARSE] Handling StateSyncMessage (Type 5)... (not implemented)")
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
		// log.Println("[RLP-PARSE] -> It's a FullNodeRaptorcastRequest")
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
	var consensusMsg consensus.ConsensusMessage
	if err := rlp.Decode(bytes.NewReader(payload), &consensusMsg); err != nil {
		return fmt.Errorf("stage 2 (ConsensusMessage) decode failed: %w", err)
	}

	switch msg := consensusMsg.Message.(type) {
	case *proposal.ProposalMessage:
		// log.Printf("[RLP-PARSE]   -> IT'S A PROPOSAL! Round: %d, Epoch: %d", msg.ProposalRound, msg.ProposalEpoch)
	case *vote.VoteMessage:
		// log.Printf("[RLP-PARSE]   -> IT'S A VOTE! Round: %d, BlockID: %s", msg.Vote.Round, msg.Vote.ID.String())
	case *timeout.TimeoutMessage:
		// log.Printf("[RLP-PARSE]   -> IT'S A TIMEOUT! Round: %d, Epoch: %d", msg.TMInfo.Round, msg.TMInfo.Epoch)
	case *round_recovery.RoundRecoveryMessage:
		// log.Printf("[RLP-PARSE]   -> IT'S A ROUND RECOVERY! Round: %d, Epoch: %d", msg.Round, msg.Epoch)
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

func handleForwardedTx(payload []byte) error {
	var forwardedTxMsg forwarded_tx.ForwardedTxMessage
	if err := rlp.Decode(bytes.NewReader(payload), &forwardedTxMsg); err != nil {
		return fmt.Errorf("stage 2 (ForwardedTxMessage) decode failed: %w", err)
	}
	// log.Printf("[RLP-PARSE]   -> IT'S AN FORWARDEDTX! (Total Txs: %d)", len(forwardedTxMsg))
	// for i, tx := range forwardedTxMsg {
	// 	end := 10
	// 	if len(tx) < 10 {
	// 		end = len(tx)
	// 	}
	// 	log.Printf("[RLP-PARSE]     -> Tx[%d] (First %dB): %x...", i, end, tx[:end])
	// }
	return nil
}
