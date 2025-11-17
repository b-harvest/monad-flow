import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false, discriminatorKey: 'type' })
export class PeerDiscoveryBase {
  @Prop({ required: true })
  type: string;
}
export const PeerDiscoveryBaseSchema =
  SchemaFactory.createForClass(PeerDiscoveryBase);
