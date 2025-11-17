import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { QuorumCertificateSchema } from './quorum-certificate.schema';
import { ProposedHeaderSchema } from './proposed-header.schema';

@Schema({ _id: false })
export class ConsensusBlockHeader {
  @Prop({ required: true })
  blockRound: string;

  @Prop({ required: true })
  epoch: string;

  @Prop({ type: QuorumCertificateSchema, required: true })
  qc: any;

  @Prop({ required: true })
  author: string; // NodeID hex

  @Prop({ required: true })
  seqNum: string;

  @Prop({ required: true })
  timestampNs: string; // big.Int → string

  @Prop({ required: true })
  roundSignature: string; // hex

  @Prop({ type: [Object], default: [] })
  delayedExecutionResults: any[];

  @Prop({ type: ProposedHeaderSchema, required: true })
  executionInputs: any;

  @Prop({ required: true })
  blockBodyId: string; // hash hex

  @Prop()
  baseFee?: string;

  @Prop()
  baseFeeTrend?: string;

  @Prop()
  baseFeeMoment?: string;
}

export const ConsensusBlockHeaderSchema =
  SchemaFactory.createForClass(ConsensusBlockHeader);
