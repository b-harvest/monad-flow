import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import {
  NetworkMessageVersion,
  NetworkMessageVersionSchema,
} from './common/network-message-version.schema';
import { PeerDiscoveryBaseSchema } from './peer-discovery/peer-discovery-base.schema';
import { PingMessage } from './peer-discovery/ping.schema';
import { PongMessage } from './peer-discovery/pong.schema';
import { PeerLookupRequestMessage } from './peer-discovery/peer-lookup-request.schema';
import { PeerLookupResponseMessage } from './peer-discovery/peer-lookup-response.schema';
import { FullNodeRaptorcastRequestMessage } from './peer-discovery/fullnode-raptorcast-request.schema';
import { FullNodeRaptorcastResponseMessage } from './peer-discovery/fullnode-raptorcast-response.schema';
import { FullNodesGroupMessageBaseSchema } from './fullnodes-group/fullnodes-group-base.schema';
import { PrepareGroup } from './fullnodes-group/prepare-group.schema';
import { ConfirmGroup } from './fullnodes-group/confirm-group.schema';
import { PrepareGroupResponse } from './fullnodes-group/prepare-group-response.schema';
import {
  MonadMessage,
  MonadMessageSchema,
} from './app-message/monad-message.schema';

@Schema()
export class OutboundRouterMessageDocument extends Document {
  /**
   * Version (SerializeVersion + CompressionVersion)
   */
  @Prop({ type: NetworkMessageVersionSchema, required: true })
  version: NetworkMessageVersion;

  /**
   * Message type (uint8)
   */
  @Prop({ required: true })
  messageType: number;

  /**
   * Decoded PeerDiscoveryMessage (Union)
   */
  @Prop({ type: PeerDiscoveryBaseSchema })
  decoded?:
    | PingMessage
    | PongMessage
    | PeerLookupRequestMessage
    | PeerLookupResponseMessage
    | FullNodeRaptorcastRequestMessage
    | FullNodeRaptorcastResponseMessage;

  @Prop({ type: FullNodesGroupMessageBaseSchema })
  fullNodesGroupMessage?: PrepareGroup | PrepareGroupResponse | ConfirmGroup;

  @Prop({ type: MonadMessageSchema })
  appMessage?: MonadMessage;

  /**
   * Chunk 조립 추적
   */
  @Prop({ type: [String], index: true })
  includedChunkIds: string[];

  /**
   * DB 저장 시간
   */
  @Prop({ default: Date.now })
  createdAt: Date;
}

export const OutboundRouterMessageSchema = SchemaFactory.createForClass(
  OutboundRouterMessageDocument,
);
