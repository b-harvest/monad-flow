import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { NoEndorsementCertificateSchema } from './no-endorsement-certificate.schema';

@Schema({ _id: false })
export class FreshProposalCertificateNEC {
  @Prop({ type: NoEndorsementCertificateSchema, required: true })
  nec: any;
}
export const FreshProposalCertificateNECSchema = SchemaFactory.createForClass(
  FreshProposalCertificateNEC,
);
