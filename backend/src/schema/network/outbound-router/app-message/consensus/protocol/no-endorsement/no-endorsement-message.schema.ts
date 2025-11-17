import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class NoEndorsementMessage {
  @Prop({ required: true })
  type: 'NO_ENDORSEMENT';

  @Prop({
    type: {
      epoch: String,
      round: String,
      tipQcRound: String,
    },
    required: true,
  })
  msg: {
    epoch: string;
    round: string;
    tipQcRound: string;
  };

  @Prop({ required: true })
  sig: string; // hex
}
export const NoEndorsementMessageSchema =
  SchemaFactory.createForClass(NoEndorsementMessage);
