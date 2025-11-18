import { Injectable, Logger } from '@nestjs/common';
import { InjectModel } from '@nestjs/mongoose';
import { Model } from 'mongoose';
import { OutboundRouterMessage } from './schema/network/outbound-router/outbound-router-message.schema';
import { UDPEvent } from './common/enum-definition';
import { UDPEventType } from './common/type-definition';
import { MonadChunkPacket } from './schema/network/monad-chunk-packet.schema';

@Injectable()
export class AppService {
  private readonly logger = new Logger(AppService.name);

  constructor(
    @InjectModel(MonadChunkPacket.name)
    private readonly chunkModel: Model<MonadChunkPacket>,
    @InjectModel(OutboundRouterMessage.name)
    private readonly outboundRouterModel: Model<OutboundRouterMessage>,
  ) {}

  async createFromUDP(payload: {
    type: UDPEventType;
    data: any;
  }): Promise<void> {
    const { type, data } = payload;
    if (type === UDPEvent.MONAD_CHUNK_PACKET) {
      await this.handleMonadChunkPacket(data);
    } else {
      this.logger.warn(`Unknown UDPEvent type: ${type}`);
    }
  }

  private async handleMonadChunkPacket(data: any): Promise<void> {
    const doc = new this.chunkModel({
      network: {
        ipv4: {
          srcIp: data.Network?.Ipv4?.SrcIp,
          dstIp: data.Network?.Ipv4?.DstIp,
          protocol: data.Network?.Ipv4?.Protocol,
        },
        port: {
          srcPort: data.Network?.Port?.SrcPort,
          dstPort: data.Network?.Port?.DstPort,
        },
      },

      signature: Buffer.from(data.Signature || []),
      version: data.Version,
      flags: data.Flags,
      broadCast: data.Broadcast,
      secondaryBroadcast: data.SecondaryBroadcast,
      merkleTreeDepth: data.MerkleTreeDepth,

      epoch: data.Epoch?.toString(),
      timestampMs: data.TimestampMs,
      timestamp: new Date(data.TimestampMs),

      appMessageHash: Buffer.from(data.AppMessageHash || []),
      appMessageLen: data.AppMessageLen,

      merkleProof: data.MerkleProof,
      firstHopRecipient: Buffer.from(data.FirstHopRecipient || []),

      merkleLeafIdx: data.MerkleLeafIdx,
      reserved: data.Reserved,
      chunkId: data.ChunkID,
    });
    await doc.save();
    this.logger.log(
      `[DB] Saved MonadChunkPacket: epoch=${data.Epoch}, chunk=${data.ChunkID}, appMessageHash=${data.AppMessageHash}, timestampMs=${data.TimestampMs}ms`,
    );
  }

  async getAll(): Promise<void> {}
}
