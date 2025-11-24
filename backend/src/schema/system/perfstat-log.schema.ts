import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

class PerfMetric {
  @Prop({ required: true })
  event: string;

  @Prop()
  value?: string;

  @Prop()
  unit?: string;

  @Prop()
  metric_val?: string;

  @Prop()
  run_pct?: string;
}

@Schema({ timestamps: true })
export class PerfStatLog extends Document {
  /**
   * 실제 이벤트 발생 timestamp
   */
  @Prop({ required: true, index: true })
  timestamp: Date;

  /**
   * perf stat 내 타임스탬프 (경과 시간)
   * 문자열로 들어오기 때문에 string 유지
   */
  @Prop({ required: true })
  perf_timestamp: string;

  /**
   * 모니터링 대상 PID
   */
  @Prop({ required: true })
  pid: string;

  /**
   * perf stat metrics 배열
   */
  @Prop({ type: [Object], required: true })
  metrics: PerfMetric[];
}

export const PerfStatLogSchema = SchemaFactory.createForClass(PerfStatLog);
