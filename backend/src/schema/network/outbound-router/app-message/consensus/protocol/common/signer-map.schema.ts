import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class SignerMap {
  @Prop({ required: true })
  numBits: number;

  @Prop({ required: true })
  buf: string; // hex
}
export const SignerMapSchema = SchemaFactory.createForClass(SignerMap);
