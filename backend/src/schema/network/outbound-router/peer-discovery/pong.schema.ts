import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class PongMessage {
  @Prop({ required: true })
  type: 'PONG';

  @Prop({ required: true })
  pingId: number;

  @Prop({ required: true })
  localRecordSeq: string; // uint64 → string
}
export const PongMessageSchema = SchemaFactory.createForClass(PongMessage);
