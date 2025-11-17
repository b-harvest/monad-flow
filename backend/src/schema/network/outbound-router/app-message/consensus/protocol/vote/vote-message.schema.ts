import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class VoteMessage {
  @Prop({ required: true })
  type: 'VOTE';

  @Prop({
    type: {
      id: String, // BlockID → hex
      round: String, // uint64 → string
      epoch: String, // uint64 → string
    },
    required: true,
  })
  vote: {
    id: string;
    round: string;
    epoch: string;
  };

  @Prop({ required: true })
  sig: string; // hex signature
}
export const VoteMessageSchema = SchemaFactory.createForClass(VoteMessage);
