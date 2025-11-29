package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
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

const (
	WorkerCount = 4
	QueueSize   = 10000
)

var BackendURL = getBackendURL() + "/api/outbound-message"

type taskPayload struct {
	Combined       model.OutboundRouterCombined
	AppMessageHash string
}

var taskQueue = make(chan taskPayload, QueueSize)

var httpClient *http.Client

func init() {
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}

	httpClient = &http.Client{
		Transport: t,
		Timeout:   60 * time.Second,
	}

	for i := 0; i < WorkerCount; i++ {
		go startWorker(i)
	}
	log.Printf("Initialized %d background workers (Low Concurrency Mode)", WorkerCount)
}

func startWorker(id int) {
	for task := range taskQueue {
		if err := outboundRouterSend(task.Combined, task.AppMessageHash); err != nil {
			log.Printf("[Worker-%d] Failed finally: %v", id, err)
		}
	}
}

func outboundRouterSend(combined model.OutboundRouterCombined, appMessageHash string) error {
	captureTime := time.Now()
	jsonData, err := json.Marshal(combined)
	if err != nil {
		return fmt.Errorf("marshal combined: %v", err)
	}

	payload := map[string]interface{}{
		"type":           util.OUTBOUND_ROUTER_EVENT,
		"appMessageHash": appMessageHash,
		"data":           json.RawMessage(jsonData),
		"timestamp":      captureTime.UnixMicro(),
	}

	finalBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %v", err)
	}

	var resp *http.Response
	var reqErr error

	for i := 0; i < 3; i++ {
		resp, reqErr = httpClient.Post(BackendURL, "application/json", bytes.NewBuffer(finalBody))
		if reqErr == nil {
			break
		}
		sleepTime := time.Millisecond * 500 * time.Duration(i+1)
		time.Sleep(sleepTime)
	}

	if reqErr != nil {
		return fmt.Errorf("request failed after retries: %v", reqErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backend status: %s", resp.Status)
	}

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
			return fmt.Errorf("decode MonadMessage failed: %w", err)
		}
		combined.AppMessage = msg
	default:
		return nil
	}

	select {
	case taskQueue <- taskPayload{Combined: combined, AppMessageHash: appMessageHash}:
		return nil
	default:
		log.Printf("WARNING: Task queue full (%d). Dropping message %s", QueueSize, appMessageHash)
		return nil
	}
}

func getBackendURL() string {
	if err := godotenv.Load(); err != nil {
	}
	url := os.Getenv("BACKEND_URL")
	if url == "" {
		url = "http://localhost:3000"
	}
	return url
}
