import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { Document } from 'mongoose';

@Schema()
export class MonadExecutionLog extends Document {
  /**
   * 로그가 발생한 systemd 유닛 이름
   * (예: "monad-bft.service")
   * (_SYSTEMD_UNIT 필드)
   */
  @Prop({ required: true })
  unit: string;

  /**
   * 로그 메시지 본문
   * (MESSAGE 필드)
   */
  @Prop({ required: true })
  message: string;

  /**
   * 로그가 발생한 실제 시간 (이벤트 시간)
   * (__REALTIME_TIMESTAMP 필드)
   */
  @Prop({ required: true, index: true })
  timestamp: Date;

  /**
   * 로그를 생성한 프로세스 ID
   * (_PID 필드, Go 코드에서 추가 수집 필요)
   */
  @Prop({ required: true })
  pid?: string;
}

export const MonadExecutionLogSchema =
  SchemaFactory.createForClass(MonadExecutionLog);
