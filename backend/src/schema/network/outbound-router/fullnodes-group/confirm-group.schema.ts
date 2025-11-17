import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { PrepareGroup, PrepareGroupSchema } from './prepare-group.schema';
import {
  MonadNameRecord,
  MonadNameRecordSchema,
} from '../peer-discovery/common/monad-name-record.schema';

@Schema({ _id: false })
export class ConfirmGroup {
  @Prop({ required: true })
  type: 'CONFIRM_GROUP';

  @Prop({ type: PrepareGroupSchema, required: true })
  prepare: PrepareGroup;

  @Prop({ type: [String], required: true })
  peers: string[]; // array of NodeID hex strings

  @Prop({ type: [MonadNameRecordSchema], default: [] })
  nameRecords: MonadNameRecord[];
}

export const ConfirmGroupSchema = SchemaFactory.createForClass(ConfirmGroup);
