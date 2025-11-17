import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class ExecutionBody {
  @Prop({ type: [String], default: [] })
  transactions: string[]; // hex encoded RLP tx

  @Prop({ type: [String], default: [] })
  ommers: string[]; // placeholder (ommer is empty struct)

  @Prop({ type: [String], default: [] })
  withdrawals: string[]; // hex encoded withdrawal
}

export const ExecutionBodySchema = SchemaFactory.createForClass(ExecutionBody);
