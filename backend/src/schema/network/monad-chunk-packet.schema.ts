import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class MonadChunkPacket extends Document {
  @Prop({ type: Object })
  network?: {
    ipv4?: {
      srcIp: string;
      dstIp: string;
      protocol: string;
    };
    port?: {
      srcPort: number;
      dstPort: number;
    };
  };

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

  @Prop({ required: true })
  epoch: string;

  @Prop({ required: true })
  timestampMs: string;

  @Prop({ required: true })
  appMessageHash: string; // hex (20 bytes)

  @Prop({ required: true })
  appMessageLen: number;

  @Prop({ type: [String], default: [] })
  merkleProof: string[];

  @Prop({ required: true })
  firstHopRecipient: string; // hex (20 bytes)

  @Prop({ required: true })
  merkleLeafIdx: number;

  @Prop({ required: true })
  reserved: number;

  @Prop({ required: true, index: true })
  chunkId: number;

  @Prop({ type: String, index: true, sparse: true })
  combinedMessageId?: string;

  @Prop({ default: Date.now })
  timestamp: Date;
}

export const MonadChunkPacketSchema =
  SchemaFactory.createForClass(MonadChunkPacket);
