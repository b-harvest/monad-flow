import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { NameRecord, NameRecordSchema } from './name-record.schema';

@Schema({ _id: false })
export class MonadNameRecord {
  /**
   * NameRecord sub-structure
   */
  @Prop({ type: NameRecordSchema, required: true })
  nameRecord: NameRecord;

  /**
   * Signature (byte[] → hex string)
   */
  @Prop({ required: true })
  signature: string;
}

export const MonadNameRecordSchema =
  SchemaFactory.createForClass(MonadNameRecord);
