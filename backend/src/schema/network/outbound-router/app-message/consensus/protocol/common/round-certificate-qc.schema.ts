import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { QuorumCertificateSchema } from './quorum-certificate.schema';

@Schema({ _id: false })
export class RoundCertificateQC {
  @Prop({ type: QuorumCertificateSchema, required: true })
  qc: any;
}

export const RoundCertificateQCSchema =
  SchemaFactory.createForClass(RoundCertificateQC);
