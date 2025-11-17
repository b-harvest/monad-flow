import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class PerfschedLatencySnapshot extends Document {
  /**
   * 이 리포트가 생성/분석된 시간 (파싱 시 주입 필요)
   * (한 리포트의 모든 라인은 이 값이 동일해야 합니다)
   */
  @Prop({ required: true, index: true })
  reportTime: Date;

  /**
   * 태스크(프로세스/스레드) 이름
   * (예: "monad", "Quill_Backend")
   */
  @Prop({ required: true, index: true })
  taskName: string;

  /**
   * 태스크 ID (PID/TID)
   * (예: "4", "2052398")
   */
  @Prop({ required: true, index: true })
  taskId: string;

  /**
   * CPU에서 실제 실행된 총 시간 (ms)
   * (예: 243.653)
   */
  @Prop({ required: true })
  runtimeMs: number;

  /**
   * 스케줄링(문맥 교환) 횟수
   * (예: 745)
   */
  @Prop({ required: true })
  switchesCount: number;

  /**
   * '준비' 상태에서 '실행'까지의 평균 대기 시간 (ms)
   * (예: 0.191)
   */
  @Prop({ required: true })
  avgDelayMs: number;

  /**
   * '준비' 상태에서 '실행'까지의 최대 대기 시간 (ms)
   * (예: 137.114)
   */
  @Prop({ required: true, index: true }) // 최악의 경우를 찾기 위해 인덱싱
  maxDelayMs: number;

  /**
   * 최대 대기 시간이 시작된 시점 (초, epoch/boot)
   * (예: 2833857.812024)
   */
  @Prop({ required: true })
  maxDelayStartSec: number;

  /**
   * 최대 대기 시간이 종료된 시점 (초, epoch/boot)
   * (예: 2833857.949138)
   */
  @Prop({ required: true })
  maxDelayEndSec: number;

  /**
   * 레코드가 생성된 시간 (DB 저장 시점)
   */
  @Prop({ default: Date.now })
  createdAt: Date;
}

export const PerfschedLatencySnapshotSchema = SchemaFactory.createForClass(
  PerfschedLatencySnapshot,
);
