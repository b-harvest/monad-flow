import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import {
  MonadNameRecord,
  MonadNameRecordSchema,
} from './common/monad-name-record.schema';

@Schema({ _id: false })
export class PingMessage {
  @Prop({ required: true })
  type: 'PING';

  @Prop({ required: true })
  id: number;

  @Prop({ type: MonadNameRecordSchema })
  localNameRecord?: MonadNameRecord;
}

export const PingMessageSchema = SchemaFactory.createForClass(PingMessage);
