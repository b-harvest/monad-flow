import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false, discriminatorKey: 'type' })
export class FullNodesGroupMessageBase {
  @Prop({ required: true })
  type: string;
}

export const FullNodesGroupMessageBaseSchema = SchemaFactory.createForClass(
  FullNodesGroupMessageBase,
);
