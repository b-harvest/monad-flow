import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema({ timestamps: true })
export class TurboStatLog extends Document {
  /**
   * 측정된 timestamp
   */
  @Prop({ required: true, index: true })
  timestamp: Date;

  /**
   * 코어 번호 ("-" = aggregate)
   */
  @Prop({ required: true })
  core: string;

  /**
   * CPU 번호 ("-" = aggregate)
   */
  @Prop({ required: true })
  cpu: string;

  /**
   * 평균 클럭 속도 (MHz)
   */
  @Prop({ required: true })
  avg_mhz: number;

  /**
   * CPU Busy %
   */
  @Prop({ required: true })
  busy_pct: number;

  /**
   * Busy일 때 클럭 (MHz)
   */
  @Prop({ required: true })
  bzy_mhz: number;

  /**
   * TSC MHz
   */
  @Prop({ required: true })
  tsc_mhz: number;

  /**
   * IPC
   */
  @Prop({ required: true })
  ipc: number;

  /**
   * IRQ 횟수
   */
  @Prop({ required: true })
  irq: number;

  /**
   * Core 전력 (W)
   */
  @Prop({ required: true })
  cor_watt: number;

  /**
   * Package 전력 (W)
   * (aggregate 또는 CPU0에서 반복적으로 포함됨)
   */
  @Prop()
  pkg_watt?: number;
}

export const TurboStatLogSchema = SchemaFactory.createForClass(TurboStatLog);
