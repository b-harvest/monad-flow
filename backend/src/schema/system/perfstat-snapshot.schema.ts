import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class PerfStatSnapshot extends Document {
  /**
   * 모니터링 대상 프로세스 ID (PID)
   * 예: "2052396"
   */
  @Prop({ required: true, index: true })
  pid: string;

  /**
   * perf stat 시작 후 경과 시간 (초)
   * 예: 0.500472885
   */
  @Prop({ required: true })
  elapsedTimeSec: number;

  /**
   * 프로세스가 실제 CPU를 사용한 총 시간 (msec)
   * 예: 3157.35
   */
  @Prop({ required: true })
  taskClockMsec: number;

  /**
   * 점유 중인 평균 CPU 코어 수
   * 예: 6.315
   */
  @Prop({ required: true })
  cpusUtilized: number;

  /**
   * 문맥 교환 발생 횟수
   * 예: 15154
   */
  @Prop({ required: true })
  contextSwitchesCount: number;

  /**
   * 초당 문맥 교환 (K/sec)
   * 예: 4.800
   */
  @Prop({ required: true })
  contextSwitchesPerSecInK: number;

  /**
   * CPU 코어 간 마이그레이션 횟수
   * 예: 0
   */
  @Prop({ required: true })
  cpuMigrationsCount: number;

  /**
   * Page Fault 발생 횟수
   * 예: 0
   */
  @Prop({ required: true })
  pageFaultsCount: number;

  /**
   * 사용된 총 CPU 클럭 사이클 수
   * 예: 15130958354
   */
  @Prop({ required: true })
  cyclesCount: number;

  /**
   * 평균 유효 클럭 (GHz)
   * 예: 4.792
   */
  @Prop({ required: true })
  cyclesGhz: number;

  /**
   * 프론트엔드에서 지연된 사이클 수
   * 예: 8993751189
   */
  @Prop({ required: true })
  stalledCyclesFrontendCount: number;

  /**
   * 프론트엔드 유휴(idle) 비율 (%)
   * 예: 59.44
   */
  @Prop({ required: true })
  stalledCyclesFrontendIdlePercent: number;

  /**
   * 총 실행된 명령어 수
   * 예: 19536270459
   */
  @Prop({ required: true })
  instructionsCount: number;

  /**
   * IPC (Instruction Per Cycle, 클럭당 명령어 처리 수)
   * 예: 1.29
   */
  @Prop({ required: true })
  instructionsPerCycle: number;

  /**
   * 명령어당 지연된 사이클 수
   * 예: 0.46
   */
  @Prop({ required: true })
  stalledCyclesPerInsn: number;

  /**
   * 총 분기 명령어 수
   * 예: 3869578306
   */
  @Prop({ required: true })
  branchesCount: number;

  /**
   * 분기 예측 실패 횟수
   * 예: 82726979
   */
  @Prop({ required: true })
  branchMissesCount: number;

  /**
   * 전체 분기 대비 예측 실패 비율 (%)
   * 예: 2.14
   */
  @Prop({ required: true })
  branchMissesPercent: number;

  /**
   * L1 데이터 캐시 로드 횟수
   * 예: 7989862269
   */
  @Prop({ required: true })
  l1DCacheLoadsCount: number;

  /**
   * L1 데이터 캐시 로드 실패(miss) 횟수
   * 예: 57386909
   */
  @Prop({ required: true })
  l1DCacheLoadMissesCount: number;

  /**
   * 전체 L1 데이터 캐시 접근 대비 실패 비율 (%)
   * 예: 0.72
   */
  @Prop({ required: true })
  l1DCacheLoadMissesPercent: number;

  /**
   * 레코드가 생성된 시간
   */
  @Prop({ default: Date.now })
  createdAt: Date;
}

export const PerfStatSnapshotSchema =
  SchemaFactory.createForClass(PerfStatSnapshot);
