import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { HighTipRoundSigColTupleSchema } from './high-tip-round-sig-col-tuple.schema';
import { HighExtendWrapperSchema } from './high-extend-wrapper.schema';

@Schema({ _id: false })
export class TimeoutCertificate {
  @Prop({ required: true })
  epoch: string;

  @Prop({ required: true })
  round: string;

  @Prop({ type: [HighTipRoundSigColTupleSchema], default: [] })
  tipRounds: any[];

  @Prop({ type: HighExtendWrapperSchema })
  highExtend?: any;
}

export const TimeoutCertificateSchema =
  SchemaFactory.createForClass(TimeoutCertificate);
