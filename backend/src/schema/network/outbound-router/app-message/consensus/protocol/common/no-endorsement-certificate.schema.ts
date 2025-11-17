import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { NoEndorsementMessageInnerSchema } from './no-endorsement-message-inner.schema';

@Schema({ _id: false })
export class NoEndorsementCertificate {
  @Prop({ type: NoEndorsementMessageInnerSchema, required: true })
  msg: any;

  @Prop({ required: true })
  signatures: string; // hex-encoded []byte
}
export const NoEndorsementCertificateSchema = SchemaFactory.createForClass(
  NoEndorsementCertificate,
);
