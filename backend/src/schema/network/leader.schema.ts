import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class Leader extends Document {
  @Prop({ required: true })
  epoch: number;

  @Prop({ required: true })
  round: number;

  @Prop({ required: true })
  node_id: string;

  @Prop({ required: true })
  cert_pubkey: string;

  @Prop({ required: true })
  stake: string;

  @Prop({ default: Date.now, index: true })
  timestamp: Date;
}

export const LeaderSchema = SchemaFactory.createForClass(Leader);
