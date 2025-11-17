import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { SignerMapSchema } from './signer-map.schema';

@Schema({ _id: false })
export class SignatureCollection {
  @Prop({ type: SignerMapSchema, required: true })
  signers: any;

  @Prop({ required: true })
  sig: string; // hex BLS aggregate signature
}
export const SignatureCollectionSchema =
  SchemaFactory.createForClass(SignatureCollection);
