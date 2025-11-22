import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema({ timestamps: true })
export class OffCpuLog extends Document {
  /**
   * Off-CPU 발생 timestamp
   */
  @Prop({ required: true, index: true })
  timestamp: Date;

  /**
   * 프로세스 / 스레드 이름 (예: "monad-node", "tokio-runtime-w")
   */
  @Prop({ required: true })
  process_name: string;

  /**
   * 프로세스 ID
   */
  @Prop({ required: true, index: true })
  pid: string;

  /**
   * 스레드 ID
   */
  @Prop({ required: true, index: true })
  tid: string;

  /**
   * Off-CPU 지속 시간 (microseconds)
   */
  @Prop({ required: true })
  duration_us: number;

  /**
   * 스택 트레이스 (함수 호출 목록)
   */
  @Prop({ type: [String], required: true })
  stack: string[];
}

export const OffCpuLogSchema = SchemaFactory.createForClass(OffCpuLog);
