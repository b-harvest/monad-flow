import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class ForwardedTxMessage {
  @Prop({ type: [String], required: true })
  txs: string[];
}

export const ForwardedTxMessageSchema =
  SchemaFactory.createForClass(ForwardedTxMessage);
