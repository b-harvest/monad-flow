package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"monad-flow/model"
	"monad-flow/model/message/outbound_router"
	"monad-flow/model/message/outbound_router/fullnode_group"
	"monad-flow/model/message/outbound_router/monad"
	"monad-flow/model/message/outbound_router/peer_discovery"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/joho/godotenv"
)

var currentEpoch util.Epoch = 0
var validatorCache atomic.Value

var BaseBackendURL = getBackendURL()
var OutboundMessageAPIURL = BaseBackendURL + "/api/outbound-message"
var LeaderAPIURL = BaseBackendURL + "/api/leader"

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func init() {
	validatorCache.Store(make(map[util.Epoch][]util.Validator))
}

func HandleDecodedMessage(data []byte, appMessageHash string) error {
	var orm outbound_router.OutboundRouterMessage

	if err := rlp.Decode(bytes.NewReader(data), &orm); err != nil {
		return fmt.Errorf("decode OutboundRouterMessage failed: %w", err)
	}

	combined := model.OutboundRouterCombined{
		Version:     orm.Version,
		MessageType: orm.MessageType,
	}

	switch orm.MessageType {
	case util.PeerDiscType:
		msg, err := peer_discovery.DecodePeerDiscoveryMessage(orm.Message)
		if err != nil {
			return fmt.Errorf("decode PeerDiscovery failed: %w", err)
		}
		combined.PeerDiscovery = msg
	case util.GroupType:
		msg, err := fullnode_group.DecodeFullNodesGroupMessage(orm.Message)
		if err != nil {
			return fmt.Errorf("decode FullNodesGroup failed: %w", err)
		}
		combined.FullNodesGroup = msg
	case util.AppMsgType:
		msg, err := monad.DecodeMonadMessage(orm.Message)
		if err != nil {
			return fmt.Errorf("decode MonadMessage(AppMessage) failed: %w", err)
		}
		combined.AppMessage = msg
	default:
		return nil
	}

	return outboundRouterSend(combined, appMessageHash)
}

func outboundRouterSend(combined model.OutboundRouterCombined, appMessageHash string) error {
	captureTime := time.Now()
	jsonData, err := json.Marshal(combined)
	if err != nil {
		return fmt.Errorf("Error marshaling combined data: %v", err)
	}

	payload := map[string]interface{}{
		"type":           util.OUTBOUND_ROUTER_EVENT,
		"appMessageHash": appMessageHash,
		"data":           json.RawMessage(jsonData),
		"timestamp":      captureTime.UnixMicro(),
	}

	finalBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error marshaling final payload: %v", err)
	}

	checkAndTriggerLeaderStream(finalBody)

	resp, err := httpClient.Post(OutboundMessageAPIURL, "application/json", bytes.NewBuffer(finalBody))
	if err != nil {
		return fmt.Errorf("Failed to send to backend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Backend returned non-OK status: %s", resp.Status)
	}

	// log.Printf("Sent payload successfully. Size: %.2f KB", float64(len(finalBody))/1024.0)
	return nil
}

func checkAndTriggerLeaderStream(finalBody []byte) {
	var finalData map[string]interface{}
	if err := json.Unmarshal(finalBody, &finalData); err != nil {
		return
	}

	dataMap, ok := finalData["data"].(map[string]interface{})
	if !ok {
		return
	}
	appMessageMap, ok := dataMap["appMessage"].(map[string]interface{})
	if !ok {
		return
	}
	payloadMap, ok := appMessageMap["payload"].(map[string]interface{})
	if !ok {
		return
	}
	innerPayload, ok := payloadMap["payload"].(map[string]interface{})
	if !ok {
		return
	}
	finalInnerPayload, ok := innerPayload["payload"].(map[string]interface{})
	if !ok {
		return
	}

	parsedProposalEpoch, epochExists := finalInnerPayload["ProposalEpoch"].(float64)
	proposalEpoch := util.Epoch(uint64(parsedProposalEpoch))
	parsedProposalRound, roundExists := finalInnerPayload["ProposalRound"].(float64)
	proposalRound := util.Round(uint64(parsedProposalRound))

	if epochExists {
		if currentEpoch != proposalEpoch || currentEpoch == 0 {
			if err := cacheValidators(); err != nil {
				log.Printf("Failed to update validators cache: %v", err)
				return
			}
			currentEpoch = proposalEpoch
			if roundExists {
				go streamLeaders(proposalEpoch, proposalRound)
			}
		}
	}
}

func streamLeaders(epoch util.Epoch, startRound util.Round) {
	cacheMap := validatorCache.Load().(map[util.Epoch][]util.Validator)
	validators, ok := cacheMap[epoch]

	if !ok {
		return
	}

	const maxRounds = 50000
	const interval = 50 * time.Millisecond

	for i := 0; i < maxRounds; i++ {
		targetRound := startRound + util.Round(i)
		leader, err := util.GetLeader(uint64(targetRound), validators)
		if err != nil {
			log.Printf("Error calculating leader for Round %d: %v", targetRound, err)
			continue
		}
		sendLeaderPayload(epoch, targetRound, leader)
		time.Sleep(interval)
	}
}

func sendLeaderPayload(epoch util.Epoch, round util.Round, leader util.Validator) {
	payload := map[string]interface{}{
		"epoch":       epoch,
		"round":       round,
		"node_id":     leader.NodeID,
		"cert_pubkey": leader.CertPubkey,
		"stake":       leader.Stake,
		"timestamp":   time.Now().UnixMicro(),
	}

	finalBody, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling final payload: %v\n", err)
		return
	}

	resp, err := httpClient.Post(LeaderAPIURL, "application/json", bytes.NewBuffer(finalBody))
	if err != nil {
		fmt.Printf("Failed to send to backend: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		fmt.Printf("Backend returned non-OK status: %s\n", resp.Status)
		return
	}
}

func cacheValidators() error {
	config, err := util.LoadValidatorsConfig()
	if err != nil {
		return fmt.Errorf("failed to load validators config: %w", err)
	}
	if len(config.ValidatorSets) == 0 {
		return fmt.Errorf("no validator sets found in TOML file")
	}
	newCache := make(map[util.Epoch][]util.Validator)
	for _, vSet := range config.ValidatorSets {
		ep := util.Epoch(vSet.Epoch)
		newCache[ep] = vSet.Validators
	}
	validatorCache.Store(newCache)
	return nil
}

func getBackendURL() string {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default Backend URL")
	}
	url := os.Getenv("BACKEND_URL")
	if url == "" {
		url = "http://localhost:3000"
	}
	return url
}
