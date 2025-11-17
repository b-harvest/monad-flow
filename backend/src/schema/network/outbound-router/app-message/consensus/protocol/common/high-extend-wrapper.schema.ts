import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false, discriminatorKey: 'type' })
export class HighExtendWrapper {
  @Prop({ required: true })
  type: 'TIP' | 'QC';
}
export const HighExtendWrapperSchema =
  SchemaFactory.createForClass(HighExtendWrapper);
