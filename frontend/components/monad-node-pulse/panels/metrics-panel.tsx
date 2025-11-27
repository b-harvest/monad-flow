"use client";

import { useEffect, useMemo, useState, useSyncExternalStore } from "react";
import type { ConsensusMetrics, MonadNode } from "@/types/monad";
import type { OutboundRouterEvent } from "@/lib/api/outbound-router";
import {
  getOutboundRouterEvents,
  subscribeToOutboundRouterEvents,
} from "@/lib/storage/outbound-router-cache";

interface MetricsPanelProps {
  metrics: ConsensusMetrics;
  nodes: MonadNode[];
}

const numberFormatter = new Intl.NumberFormat("en-US");

export function MetricsPanel({ metrics, nodes }: MetricsPanelProps) {
  const leaderNode = nodes.find((node) => node.id === metrics.leaderId);
  const [hydrated, setHydrated] = useState(false);
  const outboundEvents = useSyncExternalStore(
    subscribeToOutboundRouterEvents,
    getOutboundRouterEvents,
    getOutboundRouterEvents,
  );

  useEffect(() => {
    setHydrated(true);
  }, []);

  const proposals = useMemo(() => {
    const normalized: ProposalSnapshot[] = [];
    outboundEvents.forEach((event) => {
      const snapshot = normalizeProposal(event);
      if (snapshot) {
        normalized.push(snapshot);
      }
    });
    return normalized;
  }, [outboundEvents]);

  const latestProposal =
    hydrated && proposals.length > 0 ? proposals[proposals.length - 1] : null;
  const previousProposal =
    hydrated && proposals.length > 1 ? proposals[proposals.length - 2] : null;

  const roundValue = latestProposal ? latestProposal.round : null;
  const epochValue = latestProposal ? latestProposal.epoch : null;
  const leaderValue = latestProposal?.author ?? null;
  const blockHeightValue = latestProposal ? latestProposal.seqNum : null;

  const blockDeltaNs =
    latestProposal && previousProposal
      ? latestProposal.timestampNs - previousProposal.timestampNs
      : null;
  const blockDeltaSeconds =
    blockDeltaNs && blockDeltaNs > 0 ? blockDeltaNs / 1_000_000_000 : null;
  const avgBlockTimeValue =
    typeof blockDeltaSeconds === "number" ? blockDeltaSeconds : null;
  const txPerSecond =
    blockDeltaSeconds && blockDeltaSeconds > 0
      ? (latestProposal?.txCount ?? 0) / blockDeltaSeconds
      : null;
  const tpsValue = typeof txPerSecond === "number" ? txPerSecond : null;
  const hasProposalData = hydrated && latestProposal !== null;

  return (
    <section className="hud-panel metrics-panel">
      <div className="metrics-duo">
        <div className="metrics-duo-item">
          <span className="text-label">Round</span>
          {roundValue !== null ? (
            <p className="text-display-hero">
              {numberFormatter.format(roundValue)}
            </p>
          ) : (
            <p className="text-display-hero text-placeholder">waiting…</p>
          )}
        </div>
        <div className="metrics-duo-item">
          <span className="text-label">Epoch</span>
          {epochValue !== null ? (
            <p className="text-display-hero">
              {numberFormatter.format(epochValue)}
            </p>
          ) : (
            <p className="text-display-hero text-placeholder">waiting…</p>
          )}
        </div>
      </div>

      <div className="metrics-grid">
        <MetricItem
          label="Leader"
          value={leaderValue ?? "waiting…"}
          helper={undefined}
          isPlaceholder={leaderValue === null}
        />
        <MetricItem
          label="TPS"
          value={
            typeof tpsValue === "number"
              ? numberFormatter.format(Math.round(tpsValue))
              : "waiting…"
          }
          helper="avg"
          isPlaceholder={typeof tpsValue !== "number"}
        />
        <MetricItem
          label="Block Height"
          value={
            blockHeightValue !== null
              ? numberFormatter.format(blockHeightValue)
              : "waiting…"
          }
          isPlaceholder={blockHeightValue === null}
        />
        <MetricItem
          label="Avg Block Time"
          value={
            typeof avgBlockTimeValue === "number"
              ? `${avgBlockTimeValue.toFixed(2)}s`
              : "waiting…"
          }
          helper="sliding window"
          isPlaceholder={typeof avgBlockTimeValue !== "number"}
        />
      </div>

    </section>
  );
}

interface ProposalSnapshot {
  round: number;
  epoch: number;
  seqNum: number;
  timestampNs: number;
  author?: string;
  txCount: number;
}

function normalizeProposal(event: OutboundRouterEvent): ProposalSnapshot | null {
  if (event.messageType !== 1) {
    return null;
  }
  const data = event.data as Record<string, any> | undefined | null;
  const rootTypeId = getTypeId(data);
  if (!data || rootTypeId !== 1) {
    return null;
  }
  const stageOne = data.payload;
  const stageTwo = stageOne?.payload;
  const stageThree = stageTwo?.payload;
  const messageType =
    typeof stageTwo?.messageType === "number"
      ? stageTwo?.messageType
      : Number(stageTwo?.messageType);
  if (!stageTwo || messageType !== 1 || !stageThree) {
    return null;
  }
  const blockHeader = stageThree.Tip?.BlockHeader;
  if (!blockHeader) {
    return null;
  }
  const round = Number(stageThree.ProposalRound ?? blockHeader.BlockRound);
  const epoch = Number(stageThree.ProposalEpoch ?? blockHeader.Epoch);
  const seqNum = Number(blockHeader.SeqNum ?? stageThree.ExecutionInputs?.Number);
  const timestampNsRaw =
    blockHeader.TimestampNS ??
    (stageThree.ExecutionInputs?.Timestamp
      ? Number(stageThree.ExecutionInputs.Timestamp) * 1_000_000_000
      : undefined);
  const timestampNs = Number(timestampNsRaw ?? 0);
  const txCount =
    Array.isArray(stageThree.BlockBody?.ExecutionBody?.Transactions)
      ? stageThree.BlockBody.ExecutionBody.Transactions.length
      : 0;
  if (!Number.isFinite(round) || !Number.isFinite(epoch) || !Number.isFinite(seqNum)) {
    return null;
  }
  return {
    round,
    epoch,
    seqNum,
    timestampNs: Number.isFinite(timestampNs) ? timestampNs : 0,
    author: blockHeader.Author,
    txCount,
  };
}

function getTypeId(value: Record<string, any> | undefined | null) {
  if (!value) return null;
  const raw = value.typeId ?? value.TypeID;
  if (raw === undefined) return null;
  const num = typeof raw === "number" ? raw : Number(raw);
  return Number.isFinite(num) ? num : null;
}

interface MetricItemProps {
  label: string;
  value: string;
  helper?: string;
  variant?: "default" | "success" | "danger";
  isPlaceholder?: boolean;
}

function MetricItem({
  label,
  value,
  helper,
  variant = "default",
  isPlaceholder = false,
}: MetricItemProps) {
  return (
    <div className={`metric-item ${variant}`}>
      <span className="text-label">{label}</span>
      <span
        className={`text-number ${label === "Leader" ? "truncate" : ""} ${
          isPlaceholder ? "text-placeholder" : ""
        }`}
      >
        {label === "Leader" && value.length > 18
          ? `${value.slice(0, 15)}…`
          : value}
      </span>
      {helper ? <span className="metric-helper">{helper}</span> : null}
    </div>
  );
}
