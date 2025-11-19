import { Module } from '@nestjs/common';
import { MongooseModule } from '@nestjs/mongoose';
import {
  OutboundRouterMessage,
  OutboundRouterMessageSchema,
} from '../schema/network/outbound-router-message.schema';
import {
  MonadChunkPacket,
  MonadChunkPacketSchema,
} from '../schema/network/monad-chunk-packet.schema';

@Module({
  imports: [
    MongooseModule.forFeature([
      {
        name: OutboundRouterMessage.name,
        schema: OutboundRouterMessageSchema,
        collection: 'outbound_router_messages',
      },
      {
        name: MonadChunkPacket.name,
        schema: MonadChunkPacketSchema,
        collection: 'monad_chunk_packets',
      },
    ]),
  ],
  exports: [MongooseModule],
})
export class ModelModule {}
