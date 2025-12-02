import type { OutboundRouterEvent } from "@/lib/api/outbound-router";

export interface ProposalSnapshot {
  round: number;
  epoch: number;
  seqNum: number;
  timestampNs: number;
  author?: string;
  txCount: number;
  transactions: any[];
}

type UnknownRecord = Record<string, unknown>;

const toNumber = (value: unknown): number | null => {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
};

const asRecord = (value: unknown): UnknownRecord | null => {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as UnknownRecord;
  }
  return null;
};

const getTypeId = (value: UnknownRecord | null) => {
  if (!value) {
    return null;
  }
  const raw = value.typeId ?? value.TypeID;
  return toNumber(raw);
};

export function getProposalSnapshot(
  event: OutboundRouterEvent,
): ProposalSnapshot | null {
  if (event.messageType !== 1) {
    return null;
  }
  const data = asRecord(event.data);
  if (!data || getTypeId(data) !== 1) {
    return null;
  }

  const stageOne = asRecord(data.payload);
  const stageTwo = asRecord(stageOne?.payload);
  const stageThree = asRecord(stageTwo?.payload);
  const messageType = toNumber(stageTwo?.messageType);
  if (!stageThree || messageType !== 1) {
    return null;
  }

  const tip = asRecord(stageThree.Tip);
  const blockHeader = asRecord(tip?.BlockHeader);
  if (!blockHeader) {
    return null;
  }

  const round =
    toNumber(stageThree.ProposalRound) ?? toNumber(blockHeader.BlockRound);
  const epoch =
    toNumber(stageThree.ProposalEpoch) ?? toNumber(blockHeader.Epoch);
  const executionInputs = asRecord(stageThree.ExecutionInputs);
  const seqNum =
    toNumber(executionInputs?.Number) ?? toNumber(blockHeader.SeqNum);
  const executionTimestamp = toNumber(executionInputs?.Timestamp);
  const timestampNs =
    toNumber(blockHeader.TimestampNS) ??
    (executionTimestamp !== null ? executionTimestamp * 1_000_000_000 : null);

  const blockBody = asRecord(stageThree.BlockBody);
  const executionBody = asRecord(blockBody?.ExecutionBody);
  const transactions = executionBody?.Transactions;
  const txCount = Array.isArray(transactions) ? transactions.length : 0;

  if (round === null || epoch === null || seqNum === null) {
    return null;
  }

  const resolvedTimestampNs =
    timestampNs !== null && Number.isFinite(timestampNs) ? timestampNs : 0;

  return {
    round,
    epoch,
    seqNum,
    timestampNs: resolvedTimestampNs,
    author:
      typeof blockHeader.Author === "string" ? blockHeader.Author : undefined,
    txCount,
    transactions: Array.isArray(transactions) ? transactions : [],
  };
}
