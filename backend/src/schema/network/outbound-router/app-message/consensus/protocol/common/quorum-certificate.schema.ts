import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { VoteMessageSchema } from '../vote/vote-message.schema';

@Schema({ _id: false })
export class QuorumCertificate {
  @Prop({ type: VoteMessageSchema, required: true })
  info: any; // vote.VoteMessage

  @Prop({ required: true })
  signatures: string; // raw rlp hex
}

export const QuorumCertificateSchema =
  SchemaFactory.createForClass(QuorumCertificate);
