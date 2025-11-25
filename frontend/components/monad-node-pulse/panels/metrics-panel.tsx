"use client";

import type { ConsensusMetrics, MonadNode } from "@/types/monad";

interface MetricsPanelProps {
  metrics: ConsensusMetrics;
  nodes: MonadNode[];
}

const numberFormatter = new Intl.NumberFormat("en-US");

export function MetricsPanel({ metrics, nodes }: MetricsPanelProps) {
  const leader = nodes.find((node) => node.id === metrics.leaderId);

  return (
    <section className="hud-panel metrics-panel">
      <div className="metrics-duo">
        <div className="metrics-duo-item">
          <span className="text-label">Round</span>
          <p className="text-display-hero">{numberFormatter.format(metrics.round)}</p>
        </div>
        <div className="metrics-duo-item">
          <span className="text-label">Epoch</span>
          <p className="text-display-hero">{numberFormatter.format(metrics.epoch)}</p>
        </div>
      </div>

      <div className="metrics-grid">
        <MetricItem
          label="Leader"
          value={leader?.name ?? "Unknown"}
          helper={leader?.ip}
        />
        <MetricItem
          label="TPS"
          value={numberFormatter.format(Math.round(metrics.tps))}
          helper="avg"
        />
        <MetricItem
          label="Block Height"
          value={numberFormatter.format(metrics.blockHeight)}
        />
        <MetricItem
          label="Avg Block Time"
          value={`${metrics.avgBlockTime.toFixed(2)}s`}
          helper="sliding window"
        />
      </div>

      <div className="metrics-health">
        <div className="metrics-health-head">
          <span className="text-label">Network Health</span>
          <span className="text-number">{metrics.networkHealth}%</span>
        </div>
        <div className="health-bar">
          <div
            className="health-bar-fill"
            style={{ width: `${metrics.networkHealth}%` }}
          />
        </div>
      </div>
    </section>
  );
}

interface MetricItemProps {
  label: string;
  value: string;
  helper?: string;
  variant?: "default" | "success" | "danger";
}

function MetricItem({
  label,
  value,
  helper,
  variant = "default",
}: MetricItemProps) {
  return (
    <div className={`metric-item ${variant}`}>
      <span className="text-label">{label}</span>
      <span className="text-number">{value}</span>
      {helper ? <span className="metric-helper">{helper}</span> : null}
    </div>
  );
}
