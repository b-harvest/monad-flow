"use client";

import { useEffect, useMemo, useRef } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import type { ConsensusMetrics, MonadNode } from "@/types/monad";
import { KNOWN_PEERS } from "@/lib/monad/known-peers";

interface MetricsPanelProps {
  metrics: ConsensusMetrics;
  nodes: MonadNode[];
}

const numberFormatter = new Intl.NumberFormat("en-US");

export function MetricsPanel({ metrics, nodes }: MetricsPanelProps) {
  const proposalSnapshots = useNodePulseStore(
    (state) => state.proposalSnapshots,
  );

  const { latestProposal, previousProposal } = useMemo(() => {
    const count = proposalSnapshots.length;
    return {
      latestProposal: proposalSnapshots[count - 1] ?? null,
      previousProposal: proposalSnapshots[count - 2] ?? null,
    };
  }, [proposalSnapshots]);

  const roundValue = latestProposal ? latestProposal.round : null;
  const epochValue = latestProposal ? latestProposal.epoch : null;
  const leaderValue =
    getLeaderNameFromAuthor(latestProposal?.author) ??
    (latestProposal?.author ?? null);
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

function tryDecodeAuthorToHex(author: string): string | null {
  let value = author.trim();
  if (!value) return null;

  // 1) Base64 / base64url 문자열 시도 (예: "A3Wc0tBomfecW4+cMSC1A55WlUPrclXASX7Vm8BIJ2wm")
  //    => 33바이트 압축 pubkey -> 66자리 hex
  try {
    // base64url 변형도 허용
    const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
    // 패딩 없으면 추가 시도
    const padded =
      normalized.length % 4 === 0
        ? normalized
        : normalized + "=".repeat(4 - (normalized.length % 4));
    if (/^[0-9A-Za-z+/=]+$/.test(padded)) {
      const binary =
        typeof atob === "function" ? atob(padded) : null;
      if (binary) {
        let hex = "";
        for (let i = 0; i < binary.length; i += 1) {
          const byte = binary.charCodeAt(i);
          if (byte < 0 || byte > 255) {
            hex = "";
            break;
          }
          hex += byte.toString(16).padStart(2, "0");
        }
        if (hex && hex.length >= 64) {
          return hex.toLowerCase();
        }
      }
    }
  } catch {
    // base64 decode 실패는 무시하고 다른 포맷 시도
  }

  if (value.startsWith("0x") || value.startsWith("0X")) {
    const hex = value.slice(2);
    if (/^[0-9a-fA-F]+$/.test(hex)) {
      return hex.toLowerCase();
    }
  }

  if (value.startsWith("[") && value.endsWith("]")) {
    try {
      const parsed = JSON.parse(value);
      if (Array.isArray(parsed)) {
        const parts = parsed.map((b) => {
          if (
            typeof b === "number" &&
            Number.isInteger(b) &&
            b >= 0 &&
            b <= 255
          ) {
            return b.toString(16).padStart(2, "0");
          }
          return null;
        });
        if (parts.every((p) => p !== null)) {
          return (parts as string[]).join("").toLowerCase();
        }
      }
    } catch {
      // ignore
    }
  }

  if (/^[0-9a-fA-F]+$/.test(value)) {
    return value.toLowerCase();
  }

  return null;
}

function getLeaderNameFromAuthor(author: string | undefined): string | null {
  if (!author) return null;
  const hex = tryDecodeAuthorToHex(author);
  if (!hex) return null;
  const known = KNOWN_PEERS[hex] as { name?: string } | undefined;
  if (known && typeof known.name === "string" && known.name.length > 0) {
    return known.name;
  }
  return formatNodeId(`0x${hex}`);
}

function formatNodeId(id: string) {
  if (!id) return "—";
  if (id.length <= 18) return id;
  return `${id.slice(0, 10)}…${id.slice(-6)}`;
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
