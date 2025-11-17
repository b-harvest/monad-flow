import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class TurboStatSnapshot extends Document {
  /**
   * 모든 코어의 데이터를 하나로 묶기 위한 공통 스냅샷 시간 (데이터 파싱 시 주입 필요)
   */
  @Prop({ required: true, index: true })
  snapshotTime: Date;

  /**
   * 이 라인이 전체 요약(aggregate) 라인인지 여부
   * true: Core와 CPU가 '-'인 요약 라인
   * false: 개별 코어 라인
   */
  @Prop({ required: true, index: true })
  isAggregate: boolean;

  /**
   * Core 번호 (요약 라인의 경우 '-')
   */
  @Prop({ required: true })
  core: string;

  /**
   * CPU(스레드) 번호 (요약 라인의 경우 '-')
   */
  @Prop({ required: true })
  cpu: string;

  /**
   * 평균 클럭 속도 (MHz)
   */
  @Prop({ required: true })
  avgMhz: number;

  /**
   * CPU 활성 시간 비율 (%)
   */
  @Prop({ required: true })
  busyPercent: number;

  /**
   * CPU 활성 상태일 때의 평균 클럭 속도 (MHz)
   */
  @Prop({ required: true })
  bzyMhz: number;

  /**
   * 기준 클럭 속도 (MHz)
   */
  @Prop({ required: true })
  tscMhz: number;

  /**
   * IPC (Instructions Per Cycle)
   */
  @Prop({ required: true })
  ipc: number;

  /**
   * 하드웨어 인터럽트 횟수
   */
  @Prop({ required: true })
  irqCount: number;

  /**
   * POLL 상태 횟수
   */
  @Prop({ required: true })
  pollCount: number;

  /**
   * C1 유휴 상태 횟수
   */
  @Prop({ required: true })
  c1Count: number;

  /**
   * C2 유휴 상태 횟수
   */
  @Prop({ required: true })
  c2Count: number;

  /**
   * C3 유휴 상태 횟수
   */
  @Prop({ required: true })
  c3Count: number;

  /**
   * POLL 상태 시간 비율 (%)
   */
  @Prop({ required: true })
  pollPercent: number;

  /**
   * C1 유휴 상태 시간 비율 (%)
   */
  @Prop({ required: true })
  c1Percent: number;

  /**
   * C2 유휴 상태 시간 비율 (%)
   */
  @Prop({ required: true })
  c2Percent: number;

  /**
   * C3 유휴 상태 시간 비율 (%)
   */
  @Prop({ required: true })
  c3Percent: number;

  /**
   * 개별 코어 소비 전력 (Watt)
   */
  @Prop({ required: true })
  corWatt: number;

  /**
   * CPU 패키지 전체 소비 전력 (Watt)
   * (참고: 요약 라인과 0번 코어 외에는 null일 수 있음)
   */
  @Prop({ required: false, nullable: true })
  pkgWatt?: number;

  /**
   * 레코드가 생성된 시간
   */
  @Prop({ default: Date.now })
  createdAt: Date;
}

export const TurboStatSnapshotSchema =
  SchemaFactory.createForClass(TurboStatSnapshot);
