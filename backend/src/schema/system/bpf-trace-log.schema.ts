import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

class BpfEnterData {
  @Prop()
  func_raw?: string;

  @Prop()
  func_clean?: string;

  @Prop()
  caller_raw?: string;

  @Prop()
  caller_clean?: string;

  @Prop({ type: [String] })
  args_hex?: string[];
}

class BpfExitData {
  @Prop()
  duration_ns?: string;

  @Prop()
  back_to_raw?: string;

  @Prop()
  back_to_clean?: string;

  @Prop()
  return_value?: string;
}

@Schema({ timestamps: true })
export class BpfTraceLog extends Document {
  /**
   * ENTER 또는 EXIT
   */
  @Prop({ required: true, enum: ['ENTER', 'EXIT'], index: true })
  event_type: 'ENTER' | 'EXIT';

  /**
   * 로그 타임스탬프
   */
  @Prop({ required: true })
  timestamp: Date;

  /**
   * 프로세스 ID
   */
  @Prop({ required: true, index: true })
  pid: string;

  /**
   * Thread ID
   */
  @Prop({ required: true })
  tid: string;

  /**
   * BPF trace payload
   * (ENTER or EXIT depending on event_type)
   */
  @Prop({
    type: {
      func_raw: String,
      func_clean: String,
      caller_raw: String,
      caller_clean: String,
      args_hex: [String],
      duration_ns: String,
      back_to_raw: String,
      back_to_clean: String,
      return_value: String,
    },
  })
  data: BpfEnterData | BpfExitData;
}

export const BpfTraceLogSchema = SchemaFactory.createForClass(BpfTraceLog);
