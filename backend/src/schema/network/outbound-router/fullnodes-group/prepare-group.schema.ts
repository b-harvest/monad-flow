import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class PrepareGroup {
  @Prop({ required: true })
  type: 'PREPARE_GROUP';

  @Prop({ required: true })
  validatorId: string; // NodeID (hex)

  @Prop({ required: true })
  maxGroupSize: string; // uint64 → string

  @Prop({ required: true })
  startRound: string; // uint64 → string

  @Prop({ required: true })
  endRound: string; // uint64 → string
}

export const PrepareGroupSchema = SchemaFactory.createForClass(PrepareGroup);
