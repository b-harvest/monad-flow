import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import {
  MonadVersion,
  MonadVersionSchema,
} from './common/monad-version.schema';
import { ForwardedTxMessage } from './forwarded-tx/forwarded-tx-message.schema';
import { ConsensusMessage } from './consensus/consensus-message.schema';

@Schema({ _id: false })
export class MonadMessage {
  @Prop({ type: MonadVersionSchema, required: true })
  version: MonadVersion;

  @Prop({ required: true })
  typeId: number;

  /**
   * Decoded message:
   *  - ForwardedTxMessage
   *  - ConsensusMessage
   */
  @Prop({ type: Object })
  decoded?: ForwardedTxMessage | ConsensusMessage;
}

export const MonadMessageSchema = SchemaFactory.createForClass(MonadMessage);
