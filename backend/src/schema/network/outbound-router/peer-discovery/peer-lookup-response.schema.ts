import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import {
  MonadNameRecord,
  MonadNameRecordSchema,
} from './common/monad-name-record.schema';

@Schema({ _id: false })
export class PeerLookupResponseMessage {
  @Prop({ required: true })
  type: 'PEER_LOOKUP_RESPONSE';

  @Prop({ required: true })
  lookupId: number;

  @Prop({ required: true })
  target: string; // NodeID hex

  @Prop({ type: [MonadNameRecordSchema], default: [] })
  nameRecords: MonadNameRecord[];
}

export const PeerLookupResponseSchema = SchemaFactory.createForClass(
  PeerLookupResponseMessage,
);
