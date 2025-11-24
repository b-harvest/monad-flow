import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document, Schema as MongooseSchema } from 'mongoose';

@Schema({ timestamps: true })
export class BpfTraceLog extends Document {
  /**
   * enter | exit
   */
  @Prop({ required: true, enum: ['enter', 'exit'], index: true })
  event_type: 'enter' | 'exit';

  /**
   * 함수 이름
   */
  @Prop({ required: true })
  func_name: string;

  /**
   * 프로세스 ID
   */
  @Prop({ required: true, index: true })
  pid: string;

  /**
   * 로그가 발생한 실제 시간 (이벤트 시간)
   */
  @Prop({ required: true, index: true })
  timestamp: Date;

  /**
   * exit일 때만 존재
   */
  @Prop()
  duration_ns?: string;

  /**
   * EnterData 또는 ExitData
   * Caller / Args / BackToName / ReturnValue 등이 포함됨
   */
  @Prop({ type: MongooseSchema.Types.Mixed, required: true })
  data: any;
}

export const BpfTraceLogSchema = SchemaFactory.createForClass(BpfTraceLog);
