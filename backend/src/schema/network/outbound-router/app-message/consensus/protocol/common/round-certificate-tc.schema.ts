import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { TimeoutCertificateSchema } from './timeout-certificate.schema';

@Schema({ _id: false })
export class RoundCertificateTC {
  @Prop({ type: TimeoutCertificateSchema, required: true })
  tc: any;
}

export const RoundCertificateTCSchema =
  SchemaFactory.createForClass(RoundCertificateTC);
