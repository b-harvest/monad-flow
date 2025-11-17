import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class NoEndorsementMessageInner {
  @Prop({ required: true })
  epoch: string;

  @Prop({ required: true })
  round: string;

  @Prop({ required: true })
  tipQcRound: string;
}
export const NoEndorsementMessageInnerSchema = SchemaFactory.createForClass(
  NoEndorsementMessageInner,
);
