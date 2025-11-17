import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false, discriminatorKey: 'type' })
export class ProtocolMessageBase {
  @Prop({ required: true })
  type: string;
}

export const ProtocolMessageBaseSchema =
  SchemaFactory.createForClass(ProtocolMessageBase);
