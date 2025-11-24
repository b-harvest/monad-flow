import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema({ timestamps: true })
export class SchedulerLog extends Document {
  /**
   * 샘플 발생 시간
   */
  @Prop({ required: true, index: true })
  timestamp: Date;

  /**
   * 모니터링 타겟의 메인 PID
   */
  @Prop({ required: true })
  main_pid: string;

  /**
   * 스레드 ID
   */
  @Prop({ required: true })
  tid: string;

  /**
   * 스레드 이름 (예: monad-node, tokio-runtime-w)
   */
  @Prop({ required: true })
  thread_name: string;

  /**
   * 실행 전 대기 시간 (ms)
   */
  @Prop({ required: true })
  wait_delta_ms: number;

  /**
   * 실제 CPU 실행 시간 (ms)
   */
  @Prop({ required: true })
  run_delta_ms: number;

  /**
   * 문맥 교환 횟수
   */
  @Prop({ required: true })
  ctx_switches: number;
}

export const SchedulerLogSchema = SchemaFactory.createForClass(SchedulerLog);
