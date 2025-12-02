"use client";

import { useMemo } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import type { ConsensusMetrics, MonadNode } from "@/types/monad";
import type { ProposalSnapshot } from "@/lib/monad/normalize-proposal";

interface MetricsPanelProps {
  metrics: ConsensusMetrics;
  nodes: MonadNode[];
}

const numberFormatter = new Intl.NumberFormat("en-US");

export function MetricsPanel({ metrics, nodes }: MetricsPanelProps) {
  const leaderNode = nodes.find((node) => node.id === metrics.leaderId);
  const proposalSnapshots = useNodePulseStore(
    (state) => state.proposalSnapshots,
  );
  const leaders = useNodePulseStore((state) => state.leaders);

  const { latestProposal, previousProposal } = useMemo(() => {
    const count = proposalSnapshots.length;
    return {
      latestProposal: proposalSnapshots[count - 1] ?? null,
      previousProposal: proposalSnapshots[count - 2] ?? null,
    };
  }, [proposalSnapshots]);

  const roundValue = latestProposal ? latestProposal.round : null;
  const epochValue = latestProposal ? latestProposal.epoch : null;
  const leaderValue = latestProposal?.author ?? leaderNode?.name ?? null;
  const blockHeightValue = latestProposal ? latestProposal.seqNum : null;

  const blockDeltaNs =
    latestProposal && previousProposal
      ? latestProposal.timestampNs - previousProposal.timestampNs
      : null;
  const blockDeltaSeconds =
    blockDeltaNs && blockDeltaNs > 0 ? blockDeltaNs / 1_000_000_000 : null;
  const blockHeightDelta =
    latestProposal && previousProposal
      ? latestProposal.seqNum - previousProposal.seqNum
      : null;

  const avgBlockTimeValue =
    typeof blockDeltaSeconds === "number" &&
    blockHeightDelta !== null &&
    blockHeightDelta > 0
      ? blockDeltaSeconds / blockHeightDelta
      : null;
  const txPerSecond =
    blockDeltaSeconds && blockDeltaSeconds > 0
      ? (latestProposal?.txCount ?? 0) / blockDeltaSeconds
      : null;
  const tpsValue = typeof txPerSecond === "number" ? txPerSecond : null;

  return (
    <section className="hud-panel metrics-panel h-full flex flex-col">
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

      <div className="metrics-grid flex-1 content-stretch">
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
