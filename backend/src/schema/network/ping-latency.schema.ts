import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class PingLatency extends Document {
  @Prop({ required: true })
  ip: string;

  @Prop({ required: false })
  rtt_ms: number;

  @Prop({ default: Date.now, index: true })
  timestamp: Date;
}

export const PingLatencySchema = SchemaFactory.createForClass(PingLatency);
