import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { PrepareGroup, PrepareGroupSchema } from './prepare-group.schema';

@Schema({ _id: false })
export class PrepareGroupResponse {
  @Prop({ required: true })
  type: 'PREPARE_GROUP_RESPONSE';

  @Prop({ type: PrepareGroupSchema, required: true })
  req: PrepareGroup;

  @Prop({ required: true })
  nodeId: string; // NodeID hex

  @Prop({ required: true })
  accept: boolean;
}

export const PrepareGroupResponseSchema =
  SchemaFactory.createForClass(PrepareGroupResponse);
