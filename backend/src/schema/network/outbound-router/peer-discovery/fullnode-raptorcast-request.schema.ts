import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class FullNodeRaptorcastRequestMessage {
  @Prop({ required: true })
  type: 'RAPTORCAST_REQ';
}
export const FullNodeRaptorcastRequestSchema = SchemaFactory.createForClass(
  FullNodeRaptorcastRequestMessage,
);
