import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class NameRecord {
  /**
   * IPv4 Address (from net.IP, 4 bytes)
   */
  @Prop({ required: true })
  address: string; // e.g. "192.168.0.10"

  /**
   * Port (uint16)
   */
  @Prop({ required: true })
  port: number;

  /**
   * Sequence number (uint64 → string)
   */
  @Prop({ required: true })
  seq: string;
}

export const NameRecordSchema = SchemaFactory.createForClass(NameRecord);
