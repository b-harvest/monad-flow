import { Injectable, Logger } from '@nestjs/common';
import { InjectModel } from '@nestjs/mongoose';
import { Model } from 'mongoose';
import { OutboundRouterMessage } from './schema/network/outbound-router-message.schema';
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

  async getAll(): Promise<void> {}

  async createFromUDP(payload: {
    type: UDPEventType;
    data: any;
  }): Promise<any> {
    const { type, data } = payload;
    if (type === UDPEvent.MONAD_CHUNK_PACKET) {
      return this.handleMonadChunkPacket(data);
    } else if (type === UDPEvent.OUTBOUND_ROUTER) {
      return this.handleOutboundRouter(data);
    } else {
      this.logger.log(
        `Received UDP event:\n${JSON.stringify(payload, null, 2)}`,
      );
    }
  }

  private async handleMonadChunkPacket(data: any): Promise<MonadChunkPacket> {
    this.logger.log(
      `[DB] Saving MonadChunkPacket epoch=${data.Epoch}, chunk=${data.ChunkID}, appMessageHash=${data.AppMessageHash}`,
    );
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

      signature: '0x' + Buffer.from(data.Signature || []).toString('hex'),
      version: data.Version,
      flags: data.Flags,
      broadCast: data.Broadcast,
      secondaryBroadcast: data.SecondaryBroadcast,
      merkleTreeDepth: data.MerkleTreeDepth,

      epoch: data.Epoch?.toString(),
      timestampMs: data.TimestampMs,
      timestamp: new Date(data.TimestampMs),

      appMessageHash:
        '0x' + Buffer.from(data.AppMessageHash || []).toString('hex'),
      appMessageLen: data.AppMessageLen,

      merkleProof: data.MerkleProof,
      firstHopRecipient:
        '0x' + Buffer.from(data.FirstHopRecipient || []).toString('hex'),

      merkleLeafIdx: data.MerkleLeafIdx,
      reserved: data.Reserved,
      chunkId: data.ChunkID,
    });
    return doc.save();
  }

  private async handleOutboundRouter(
    data: any,
  ): Promise<OutboundRouterMessage> {
    this.logger.log(
      `[DB] Saving OutboundRouterMessage messageType=${data.messageType}`,
    );
    const doc = new this.outboundRouterModel({
      version: data.version,
      messageType: data.messageType,
      data:
        data.peerDiscovery || data.fullNodesGroup || data.appMessage || null,
      appMessageHash: data.appMessageHash,
    });
    return doc.save();
  }
}
