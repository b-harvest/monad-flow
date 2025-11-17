import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { NoTipCertificateSchema } from './no-tip-certificate.schema';

@Schema({ _id: false })
export class FreshProposalCertificateNoTip {
  @Prop({ type: NoTipCertificateSchema, required: true })
  noTip: any;
}
export const FreshProposalCertificateNoTipSchema = SchemaFactory.createForClass(
  FreshProposalCertificateNoTip,
);
