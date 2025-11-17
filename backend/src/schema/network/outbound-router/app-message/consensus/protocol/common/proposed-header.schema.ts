import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';

@Schema({ _id: false })
export class ProposedHeader {
  @Prop({ required: true })
  ommersHash: string; // hex

  @Prop({ required: true })
  beneficiary: string; // hex address

  @Prop({ required: true })
  transactionsRoot: string; // hex

  @Prop({ required: true })
  difficulty: string; // uint64 → string

  @Prop({ required: true })
  number: string;

  @Prop({ required: true })
  gasLimit: string;

  @Prop({ required: true })
  timestamp: string;

  @Prop({ required: true })
  extraData: string; // 32 bytes hex

  @Prop({ required: true })
  mixHash: string;

  @Prop({ required: true })
  nonce: string; // 8 bytes hex

  @Prop({ required: true })
  baseFeePerGas: string;

  @Prop({ required: true })
  withdrawalsRoot: string;

  @Prop({ required: true })
  blobGasUsed: string;

  @Prop({ required: true })
  excessBlobGas: string;

  @Prop({ required: true })
  parentBeaconBlockRoot: string;

  @Prop()
  requestsHash?: string; // optional
}

export const ProposedHeaderSchema =
  SchemaFactory.createForClass(ProposedHeader);
