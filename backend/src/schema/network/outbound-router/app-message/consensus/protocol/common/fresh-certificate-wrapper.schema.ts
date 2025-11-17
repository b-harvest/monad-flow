import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false, discriminatorKey: 'type' })
export class FreshProposalCertificateWrapper {
  @Prop({ required: true })
  type: 'NEC' | 'NOTIP';
}

export const FreshProposalCertificateWrapperSchema =
  SchemaFactory.createForClass(FreshProposalCertificateWrapper);
