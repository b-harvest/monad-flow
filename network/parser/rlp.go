package parser

import (
	"bytes"
	"fmt"
	"log"
	"monad-flow/model"
	"monad-flow/model/message/outbound_router"
	"monad-flow/model/message/outbound_router/fullnode_group"
	"monad-flow/model/message/outbound_router/monad"
	"monad-flow/model/message/outbound_router/peer_discovery"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

func HandleDecodedMessage(data []byte, appMessageHash string) error {
	log.Println("new Decoded Message IN!!")
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
	return nil // websocketOutboundRouterSend(client, clientMutex, combined, appMessageHash)
}
