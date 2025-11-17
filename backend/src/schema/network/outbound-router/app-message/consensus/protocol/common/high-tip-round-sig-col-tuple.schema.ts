import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { SignatureCollectionSchema } from './signature-collection.schema';

@Schema({ _id: false })
export class HighTipRoundSigColTuple {
  @Prop({ required: true })
  highQcRound: string;

  @Prop({ required: true })
  highTipRound: string;

  @Prop({ type: SignatureCollectionSchema, required: true })
  sigs: any;
}

export const HighTipRoundSigColTupleSchema = SchemaFactory.createForClass(
  HighTipRoundSigColTuple,
);
