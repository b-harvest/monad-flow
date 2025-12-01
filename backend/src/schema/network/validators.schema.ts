import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class Validators extends Document {
  @Prop({ required: true, unique: true })
  epoch: number;

  @Prop({ type: Object })
  data?: {
    node_id: string;
    stake: number;
    cert_pubkey: string;
  }[];
}

export const ValidatorsSchema = SchemaFactory.createForClass(Validators);
