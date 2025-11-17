import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { ExecutionBodySchema } from '../common/execution-body.schema';
import { ConsensusTipSchema } from '../common/consensus-tip.schema';
import { TimeoutCertificateSchema } from '../common/timeout-certificate.schema';

@Schema({ _id: false })
export class ProposalMessage {
  @Prop({ required: true })
  type: 'PROPOSAL';

  @Prop({ required: true })
  proposalRound: string;

  @Prop({ required: true })
  proposalEpoch: string;

  @Prop({ type: ConsensusTipSchema })
  tip: any;

  @Prop({
    type: {
      executionBody: ExecutionBodySchema,
    },
  })
  blockBody: any;

  @Prop({ type: TimeoutCertificateSchema })
  lastRoundTc?: any;
}
export const ProposalMessageSchema =
  SchemaFactory.createForClass(ProposalMessage);
