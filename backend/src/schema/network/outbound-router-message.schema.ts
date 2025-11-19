import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class OutboundRouterMessage extends Document {
  @Prop({ type: Object, required: true })
  version: any;

  @Prop({ required: true })
  messageType: number;

  @Prop({ type: Object })
  data?: any;

  @Prop({ required: false, index: true })
  appMessageHash?: string;

  @Prop({ default: Date.now })
  createdAt: Date;
}

export const OutboundRouterMessageSchema = SchemaFactory.createForClass(
  OutboundRouterMessage,
);
