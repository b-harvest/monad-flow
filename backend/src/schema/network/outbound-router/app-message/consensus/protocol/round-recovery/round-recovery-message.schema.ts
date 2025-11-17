import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { TimeoutCertificateSchema } from '../common/timeout-certificate.schema';

@Schema({ _id: false })
export class RoundRecoveryMessage {
  @Prop({ required: true })
  type: 'ROUND_RECOVERY';

  @Prop({ required: true })
  round: string;

  @Prop({ required: true })
  epoch: string;

  @Prop({ type: TimeoutCertificateSchema })
  tc: any;
}
export const RoundRecoveryMessageSchema =
  SchemaFactory.createForClass(RoundRecoveryMessage);
