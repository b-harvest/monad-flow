import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { ConsensusBlockHeaderSchema } from './consensus-block-header.schema';
import { FreshProposalCertificateWrapperSchema } from './fresh-certificate-wrapper.schema';

@Schema({ _id: false })
export class ConsensusTip {
  @Prop({ type: ConsensusBlockHeaderSchema, required: true })
  blockHeader: any;

  @Prop({ required: true })
  signature: string; // hex

  @Prop({ type: FreshProposalCertificateWrapperSchema })
  freshCertificate?: any;
}

export const ConsensusTipSchema = SchemaFactory.createForClass(ConsensusTip);
