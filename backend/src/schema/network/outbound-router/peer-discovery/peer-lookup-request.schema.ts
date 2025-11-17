import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class PeerLookupRequestMessage {
  @Prop({ required: true })
  type: 'PEER_LOOKUP_REQUEST';

  @Prop({ required: true })
  lookupId: number;

  @Prop({ required: true })
  target: string; // NodeID []byte → hex string

  @Prop({ required: true })
  openDiscovery: boolean;
}
export const PeerLookupRequestSchema = SchemaFactory.createForClass(
  PeerLookupRequestMessage,
);
