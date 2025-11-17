import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { HighExtendWrapperSchema } from '../common/high-extend-wrapper.schema';
import { RoundCertificateWrapperSchema } from '../common/round-certificate-wrapper.schema';

@Schema({ _id: false })
export class TimeoutMessage {
  @Prop({ required: true })
  type: 'TIMEOUT';

  @Prop({
    type: {
      epoch: String,
      round: String,
      highQcRound: String,
      highTipRound: String,
    },
    required: true,
  })
  tmInfo: {
    epoch: string;
    round: string;
    highQcRound: string;
    highTipRound: string;
  };

  @Prop({ required: true })
  timeoutSignature: string; // hex

  @Prop({ type: HighExtendWrapperSchema })
  highExtend: any;

  @Prop({ type: RoundCertificateWrapperSchema })
  lastRoundCertificate?: any;
}
export const TimeoutMessageSchema =
  SchemaFactory.createForClass(TimeoutMessage);
