import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class OffcputimeSnapshot extends Document {
  /**
   * 이 샘플을 수집한 시간 (데이터 파싱 시 주입 필요)
   * (예: `offcputime` 1초간 실행했다면, 그 시작 시간)
   */
  @Prop({ required: true, index: true })
  sampleTime: Date;

  /**
   * 모니터링 대상이었던 메인 프로세스 ID (PID)
   * (예: "2052396")
   */
  @Prop({ required: true, index: true })
  targetPid: string;

  /**
   * 실제 대기가 발생한 스레드(TID) 또는 프로세스(PID) ID
   * (예: "2052396", "2052721", "2052716")
   */
  @Prop({ required: true, index: true })
  tid: string;

  /**
   * 스레드/프로세스의 이름
   * (예: "monad", "worker 2", "Quill_Backend")
   */
  @Prop({ required: true })
  threadName: string;

  /**
   * 이 스택으로 인해 대기한 총 시간 (microseconds)
   * (로그 마지막의 집계 값, 예: 86911)
   */
  @Prop({ required: true })
  offCpuTimeUs: number;

  /**
   * 대기 원인을 보여주는 스택 트레이스 (함수 목록)
   */
  @Prop({ type: [String], required: true })
  stack: string[];

  /**
   * 레코드가 생성된 시간
   */
  @Prop({ default: Date.now })
  createdAt: Date;
}

export const OffcputimeSnapshotSchema =
  SchemaFactory.createForClass(OffcputimeSnapshot);
