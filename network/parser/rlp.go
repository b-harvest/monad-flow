package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

var BaseBackendURL = getBackendURL()
var OutboundMessageAPIURL = BaseBackendURL + "/api/outbound-message"
var ValidatorsAPIURL = BaseBackendURL + "/api/validators"

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func validatorsSend() error {
	config, err := util.LoadValidatorsConfig()
	if err != nil {
		return fmt.Errorf("failed to load validators config: %w", err)
	}
	if len(config.ValidatorSets) == 0 {
		return fmt.Errorf("no validator sets found in TOML file")
	}

	for _, currentSet := range config.ValidatorSets {
		payload := map[string]interface{}{
			"epoch":      currentSet.Epoch,
			"validators": currentSet.Validators,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("Error marshaling final payload: %v", err)
		}

		resp, err := httpClient.Post(ValidatorsAPIURL, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			return fmt.Errorf("Failed to send to backend: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Backend returned non-OK status: %s", resp.Status)
		}
		// log.Printf("Successfully sent Validators (Epoch %d) to API endpoint: %s", currentSet.Epoch, ValidatorsAPIURL)
	}
	return nil
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

	checkAndCacheProposalEpoch(finalBody)

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

func checkAndCacheProposalEpoch(finalBody []byte) {
	var finalData map[string]interface{}
	if err := json.Unmarshal(finalBody, &finalData); err != nil {
		log.Printf("Warning: Could not unmarshal finalBody for epoch check: %v", err)
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

	proposalEpoch, exists := finalInnerPayload["ProposalEpoch"].(float64)
	newEpoch := util.Epoch(uint64(proposalEpoch))
	if exists {
		if currentEpoch != newEpoch {
			currentEpoch = newEpoch
			if err := validatorsSend(); err != nil {
				log.Printf("Failed to send validators data to API: %v", err)
			}
		}
	}
}
