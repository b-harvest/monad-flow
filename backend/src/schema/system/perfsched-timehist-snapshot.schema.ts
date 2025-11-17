import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class PerfschedTimehistSnapshot extends Document {
  /**
   * 이벤트가 발생한 시스템 시간 (초)
   * (예: 2833857.949145)
   */
  @Prop({ required: true, index: true })
  timeSec: number;

  /**
   * 이벤트가 발생한 CPU 코어 번호
   * (예: 6)
   */
  @Prop({ required: true })
  cpu: number;

  /**
   * 태스크(프로세스/스레드) 이름
   * (예: "monad", "Quill_Backend")
   */
  @Prop({ required: true, index: true })
  taskName: string;

  /**
   * 스레드 ID (TID)
   * (예: "2052721", "2052398")
   */
  @Prop({ required: true, index: true })
  tid: string;

  /**
   * 메인 프로세스 ID (PID)
   * (예: "2052396")
   */
  @Prop({ required: true, index: true })
  pid: string;

  /**
   * 대기 시간 (msec)
   * (예: 0.000)
   */
  @Prop({ required: true })
  waitTimeMs: number;

  /**
   * 스케줄링 지연 시간 (msec) - Ready에서 Running까지
   * (예: 137.114)
   */
  @Prop({ required: true, index: true }) // 가장 중요한 지표이므로 인덱싱
  schDelayMs: number;

  /**
   * 실제 실행된 시간 (msec)
   * (예: 0.006)
   */
  @Prop({ required: true })
  runTimeMs: number;

  /**
   * 레코드가 생성된 시간 (DB 저장 시점)
   */
  @Prop({ default: Date.now })
  createdAt: Date;
}

export const PerfschedTimehistSnapshotSchema = SchemaFactory.createForClass(
  PerfschedTimehistSnapshot,
);
