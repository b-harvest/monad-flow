import { Injectable, Logger } from '@nestjs/common';
import { InjectModel } from '@nestjs/mongoose';
import { Model } from 'mongoose';
import { OutboundRouterMessage } from './schema/network/outbound-router/outbound-router-message.schema';

@Injectable()
export class AppService {
  private readonly logger = new Logger(AppService.name);

  constructor(
    @InjectModel(OutboundRouterMessage.name)
    private readonly outboundRouterModel: Model<OutboundRouterMessage>,
  ) {}

  async createFromUDP(payload: any): Promise<void> {
    this.logger.log(`Received UDP event: ${JSON.stringify(payload)}`);
    const doc = new this.outboundRouterModel({
      version: {
        serializeVersion: 1,
        compressionVersion: 0,
      },
      messageType: 999, // UDP Logging MessageType (임의)
      decoded: null,
      fullNodesGroupMessage: null,
      appMessage: null,
      includedChunkIds: [],
    });

    await doc.save();
  }

  async getAll(): Promise<void> {}
}
