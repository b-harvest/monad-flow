import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class NetworkMessageVersion {
  @Prop({ required: true })
  serializeVersion: number;

  @Prop({ required: true })
  compressionVersion: number;
}

export const NetworkMessageVersionSchema = SchemaFactory.createForClass(
  NetworkMessageVersion,
);
