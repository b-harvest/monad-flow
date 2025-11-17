import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false, discriminatorKey: 'type' })
export class RoundCertificateWrapper {
  @Prop({ required: true })
  type: 'QC' | 'TC';
}

export const RoundCertificateWrapperSchema = SchemaFactory.createForClass(
  RoundCertificateWrapper,
);
