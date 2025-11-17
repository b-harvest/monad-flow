import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { HighTipRoundSigColTupleSchema } from './high-tip-round-sig-col-tuple.schema';
import { QuorumCertificateSchema } from './quorum-certificate.schema';

@Schema({ _id: false })
export class NoTipCertificate {
  @Prop({ required: true })
  epoch: string;

  @Prop({ required: true })
  round: string;

  @Prop({ type: [HighTipRoundSigColTupleSchema], default: [] })
  tipRounds: any[];

  @Prop({ type: QuorumCertificateSchema })
  highQc?: any;
}

export const NoTipCertificateSchema =
  SchemaFactory.createForClass(NoTipCertificate);
