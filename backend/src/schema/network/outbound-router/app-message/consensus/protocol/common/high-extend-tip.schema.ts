import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { ConsensusTipSchema } from './consensus-tip.schema';

@Schema({ _id: false })
export class HighExtendTip {
  @Prop({ type: ConsensusTipSchema, required: true })
  tip: any;

  @Prop()
  voteSignature?: string; // hex
}
export const HighExtendTipSchema = SchemaFactory.createForClass(HighExtendTip);
