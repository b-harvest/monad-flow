import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class MonadVersion {
  @Prop({ required: true })
  protocolVersion: number;

  @Prop({ required: true })
  clientVersionMajor: number;

  @Prop({ required: true })
  clientVersionMinor: number;

  @Prop({ required: true })
  hashVersion: number;

  @Prop({ required: true })
  serializeVersion: number;
}

export const MonadVersionSchema = SchemaFactory.createForClass(MonadVersion);
