import { Prop, Schema, SchemaFactory } from '@nestjs/mongoose';
import { ProtocolMessageBaseSchema } from './protocol/protocol-message-base.schema';
import { VoteMessage } from './protocol/vote/vote-message.schema';
import { TimeoutMessage } from './protocol/timeout/timeout-message.schema';
import { RoundRecoveryMessage } from './protocol/round-recovery/round-recovery-message.schema';
import { ProposalMessage } from './protocol/proposal/proposal-message.schema';
import { NoEndorsementMessage } from './protocol/no-endorsement/no-endorsement-message.schema';
import { AdvanceRoundMessage } from './protocol/advanced-round/advanced-round-message.schema';

@Schema({ _id: false })
export class ConsensusMessage {
  @Prop({ required: true })
  version: number;

  @Prop({ type: ProtocolMessageBaseSchema })
  protocolMessage:
    | VoteMessage
    | TimeoutMessage
    | RoundRecoveryMessage
    | ProposalMessage
    | NoEndorsementMessage
    | AdvanceRoundMessage;
}

export const ConsensusMessageSchema =
  SchemaFactory.createForClass(ConsensusMessage);
