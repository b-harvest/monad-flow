import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class MonadChunkPacket extends Document {
  /**
   * Raw metadata from capture
   */
  @Prop({ type: Object })
  network?: {
    ipv4?: {
      srcIp: string;
      dstIp: string;
      protocol: string;
    };
    tcp?: {
      srcPort: number;
      dstPort: number;
    };
    udp?: {
      srcPort: number;
      dstPort: number;
    };
  };

  /**
   * Sender signature (65 bytes)
   */
  @Prop({ required: true })
  signature: string; // hex string

  @Prop({ required: true })
  version: number;

  @Prop({ required: true })
  flags: number;

  @Prop({ required: true })
  broadCast: boolean;

  @Prop({ required: true })
  secondaryBroadcast: boolean;

  @Prop({ required: true })
  merkleTreeDepth: number;

  /**
   * Epoch and timestamp (u64 → string)
   */
  @Prop({ required: true })
  epoch: string;

  @Prop({ required: true })
  timestampMs: string;

  /**
   * AppMessage metadata
   */
  @Prop({ required: true })
  appMessageHash: string; // hex (20 bytes)

  @Prop({ required: true })
  appMessageLen: number;

  /**
   * Merkle proof hashes
   */
  @Prop({ type: [String], default: [] })
  merkleProof: string[];

  /**
   * Chunk-level info
   */
  @Prop({ required: true })
  firstHopRecipient: string; // hex (20 bytes)

  @Prop({ required: true })
  merkleLeafIdx: number;

  @Prop({ required: true })
  reserved: number;

  @Prop({ required: true, index: true })
  chunkId: number;

  /**
   * 이 패킷이 combine/조립되어 만들어진 higher-level message 추적용
   */
  @Prop({ type: String, index: true, sparse: true })
  combinedMessageId?: string;

  /**
   * DB 저장 시간
   */
  @Prop({ default: Date.now })
  createdAt: Date;
}

export const MonadChunkPacketSchema =
  SchemaFactory.createForClass(MonadChunkPacket);
