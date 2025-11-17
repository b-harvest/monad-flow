import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { RoundCertificateWrapperSchema } from '../common/round-certificate-wrapper.schema';

@Schema({ _id: false })
export class AdvanceRoundMessage {
  @Prop({ required: true })
  type: 'ADVANCE_ROUND';

  @Prop({ type: RoundCertificateWrapperSchema, required: true })
  lastRoundCertificate: any;
}
export const AdvanceRoundMessageSchema =
  SchemaFactory.createForClass(AdvanceRoundMessage);
