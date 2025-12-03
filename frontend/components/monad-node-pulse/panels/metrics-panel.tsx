"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import type { ConsensusMetrics, MonadNode } from "@/types/monad";
import type { ProposalSnapshot } from "@/lib/monad/normalize-proposal";
import { KNOWN_PEERS } from "@/lib/monad/known-peers";
import type { LeaderEvent } from "@/lib/api/leader";
import { fetchLeaderSchedule } from "@/lib/api/leader";

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
  const [currentLeaderEvent, setCurrentLeaderEvent] =
    useState<LeaderEvent | null>(null);

  const { latestProposal, previousProposal } = useMemo(() => {
    const count = proposalSnapshots.length;
    return {
      latestProposal: proposalSnapshots[count - 1] ?? null,
      previousProposal: proposalSnapshots[count - 2] ?? null,
    };
  }, [proposalSnapshots]);

  useEffect(() => {
    if (!latestProposal) {
      setCurrentLeaderEvent(null);
      return;
    }
    const baseRound = latestProposal.round;
    let cancelled = false;

    fetchLeaderSchedule(baseRound, 5)
      .then((events) => {
        if (cancelled) return;
        const current =
          events.find(
            (event) => event.round === baseRound,
          ) ?? null;
        setCurrentLeaderEvent(current);
      })
      .catch(() => {
        if (cancelled) return;
        setCurrentLeaderEvent(null);
      });

    return () => {
      cancelled = true;
    };
  }, [latestProposal]);

  const roundValue = latestProposal ? latestProposal.round : null;
  const epochValue = latestProposal ? latestProposal.epoch : null;
  const leaderValue =
    (currentLeaderEvent &&
      getLeaderNameFromEvent(currentLeaderEvent)) ??
    latestProposal?.author ??
    leaderNode?.name ??
    null;
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
          <p
            className={`text-display-hero ${
              roundValue === null ? "text-placeholder" : ""
            }`}
          >
            {roundValue !== null ? (
              <AnimatedNumber
                value={roundValue}
                format={numberFormatter}
              />
            ) : (
              "waiting…"
            )}
          </p>
        </div>
        <div className="metrics-duo-item">
          <span className="text-label">Epoch</span>
          <p
            className={`text-display-hero ${
              epochValue === null ? "text-placeholder" : ""
            }`}
          >
            {epochValue !== null ? (
              <AnimatedNumber
                value={epochValue}
                format={numberFormatter}
              />
            ) : (
              "waiting…"
            )}
          </p>
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
            typeof tpsValue === "number" ? (
              <AnimatedNumber
                value={Math.round(tpsValue)}
                format={numberFormatter}
              />
            ) : (
              "waiting…"
            )
          }
          helper="avg"
          isPlaceholder={typeof tpsValue !== "number"}
        />
        <MetricItem
          label="Block Height"
          value={
            blockHeightValue !== null
              ? (
                  <AnimatedNumber
                    value={blockHeightValue}
                    format={numberFormatter}
                  />
                )
              : "waiting…"
          }
          isPlaceholder={blockHeightValue === null}
        />
        <MetricItem
          label="Avg Block Time"
          value={
            typeof avgBlockTimeValue === "number"
              ? (
                  <AnimatedNumber
                    value={avgBlockTimeValue}
                    format={(v) => `${v.toFixed(2)}s`}
                  />
                )
              : "waiting…"
          }
          helper="sliding window"
          isPlaceholder={typeof avgBlockTimeValue !== "number"}
        />
      </div>
    </section>
  );
}

function getLeaderNameFromEvent(event: LeaderEvent): string {
  const rawId = event.node_id.startsWith("0x")
    ? event.node_id.slice(2)
    : event.node_id;
  const known = KNOWN_PEERS[rawId] as
    | { name?: string }
    | undefined;
  if (known && typeof known.name === "string" && known.name.length > 0) {
    return known.name;
  }
  return formatNodeId(event.node_id);
}

function formatNodeId(nodeId: string) {
  if (!nodeId) return "—";
  if (nodeId.length <= 18) return nodeId;
  return `${nodeId.slice(0, 10)}…${nodeId.slice(-6)}`;
}

interface MetricItemProps {
  label: string;
  value: React.ReactNode;
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
        {label === "Leader" && typeof value === "string"
          ? value.length > 18
            ? (
                <span
                  key={value}
                  className="leader-text-swap"
                >
                  {`${value.slice(0, 15)}…`}
                </span>
              )
            : (
                <span
                  key={value}
                  className="leader-text-swap"
                >
                  {value}
                </span>
              )
          : value}
      </span>
      {helper ? <span className="metric-helper">{helper}</span> : null}
    </div>
  );
}

interface AnimatedNumberProps {
  value: number;
  format?: Intl.NumberFormat | ((value: number) => string);
}

function AnimatedNumber({ value, format }: AnimatedNumberProps) {
  const prevRef = useRef<string | null>(null);
  const formatted =
    typeof format === "function"
      ? format(value)
      : format instanceof Intl.NumberFormat
        ? format.format(value)
        : String(value);
  const prev = prevRef.current;
  useEffect(() => {
    prevRef.current = formatted;
  }, [formatted]);

  const prevChars = typeof prev === "string" ? prev.split("") : [];
  const currChars = formatted.split("");
  const maxLen = Math.max(prevChars.length, currChars.length);

  const paddedPrev = Array.from({ length: maxLen }, (_, i) => {
    const idx = prevChars.length - maxLen + i;
    return idx >= 0 ? prevChars[idx] : " ";
  });
  const paddedCurr = Array.from({ length: maxLen }, (_, i) => {
    const idx = currChars.length - maxLen + i;
    return idx >= 0 ? currChars[idx] : " ";
  });

  return (
    <span className="metric-digits">
      {paddedCurr.map((char, index) => {
        const prevChar = paddedPrev[index];
        const changed = prevChar !== char;
        const key = `${index}-${char}`;
        return (
          <span
            key={key}
            className={`metric-digit ${
              changed ? "metric-digit-changed" : ""
            }`}
          >
            {char}
          </span>
        );
      })}
    </span>
  );
}
