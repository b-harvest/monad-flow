import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class FullNodeRaptorcastResponseMessage {
  @Prop({ required: true })
  type: 'RAPTORCAST_RESP';
}
export const FullNodeRaptorcastResponseSchema = SchemaFactory.createForClass(
  FullNodeRaptorcastResponseMessage,
);
