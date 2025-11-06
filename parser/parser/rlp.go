package parser

import (
	"bytes"
	"fmt"
	"log"
	"monad-flow/model/message"
	"monad-flow/model/message/monad"
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
	var outboundRouterMsg message.OutboundRouterMessage
	if err := rlp.Decode(bytes.NewReader(data), &outboundRouterMsg); err != nil {
		return fmt.Errorf("stage 1 (container) decode failed: %w", err)
	}

	switch outboundRouterMsg.Message.TypeID {
	case util.ConsensusMsgType:
		return handleConsensusMessage(outboundRouterMsg.Message.Payload)
	case util.BlockSyncRequestMsgType:
		log.Println("[RLP-PARSE] Handling BlockSyncRequest (Type 2)... (not implemented)")
	case util.BlockSyncResponseMsgType:
		log.Println("[RLP-PARSE] Handling BlockSyncResponse (Type 3)... (not implemented)")
	case util.ForwardedTxMsgType:
		log.Println("[RLP-PARSE] Handling ForwardedTx (Type 4)... (not implemented)")
	case util.AdvanceRoundMsgType:
		log.Println("[RLP-PARSE] Handling StateSyncMessage (Type 5)... (not implemented)")
	default:
		return fmt.Errorf("unknown MonadMessage TypeID: %d", outboundRouterMsg.Message.TypeID)
	}

	return nil
}

func handleConsensusMessage(payload []byte) error {
	// log.Println("[RLP-PARSE] Handling ConsensusMessage (Type 1)...")

	var consensusMsg monad.ConsensusMessage
	if err := rlp.Decode(bytes.NewReader(payload), &consensusMsg); err != nil {
		return fmt.Errorf("stage 2 (ConsensusMessage) decode failed: %w", err)
	}

	switch msg := consensusMsg.Message.(type) {
	case *proposal.ProposalMessage:
		log.Printf("[RLP-PARSE]   -> IT'S A PROPOSAL! Round: %d, Epoch: %d", msg.ProposalRound, msg.ProposalEpoch)
	case *vote.VoteMessage:
		log.Printf("[RLP-PARSE]   -> IT'S A VOTE! Round: %d, BlockID: %s", msg.Vote.Round, msg.Vote.ID.String())
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
		log.Printf("[RLP-PARSE]   -> IT'S AN ADVANCE ROUND! QC round: %d", arm.QC.Info.Round)

	case *common.RoundCertificateTC:
		log.Printf("[RLP-PARSE]   -> IT'S AN ADVANCE ROUND! TC round: %d", arm.TC.Round)
	}
	return nil
}
