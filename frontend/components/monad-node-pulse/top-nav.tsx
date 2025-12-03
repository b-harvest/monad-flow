"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import type { PlaybackState } from "@/types/monad";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";

interface CommandNavProps {
  connectionStatus: "connected" | "degraded" | "lost";
  playback: PlaybackState;
  onPlaybackChange: (patch: Partial<PlaybackState>) => void;
  processTelemetryVisible: boolean;
  onToggleProcessTelemetry: () => void;
  historicalLoading: boolean;
  historicalError: string | null;
  onHistoricalFetch: (range: { from: number; to: number }) => void;
}

const NODE_TIME_FORMATTER = new Intl.DateTimeFormat("en-GB", {
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hour12: false,
  timeZone: "UTC",
});

const SPEED_OPTIONS: PlaybackState["speed"][] = [0.25, 0.5, 1, 2, 4];

const formatDatetimeLocal = (timestamp: number) => {
  if (!Number.isFinite(timestamp)) return "";
  const date = new Date(timestamp);
  const offsetMs = date.getTimezoneOffset() * 60000;
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
};

const parseDatetimeLocal = (value: string, fallback: number) => {
  const parsed = Date.parse(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

export function CommandNav({
  connectionStatus,
  playback,
  onPlaybackChange,
  processTelemetryVisible,
  onToggleProcessTelemetry,
  historicalError,
  historicalLoading,
  onHistoricalFetch,
}: CommandNavProps) {
  const localIp = useNodePulseStore((state) => state.localNodeIp);
  const setLocalNodeIp = useNodePulseStore((state) => state.setLocalNodeIp);
  const ensureLocalNode = useNodePulseStore((state) => state.ensureLocalNode);
  const lastEventTimestamp = useNodePulseStore(
    (state) => state.lastEventTimestamp,
  );
  const [value, setValue] = useState(localIp);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setValue(localIp);
  }, [localIp]);

  useEffect(() => {
    setMounted(true);
  }, []);

  const nodeTimeLabel =
    mounted && lastEventTimestamp > 0
      ? NODE_TIME_FORMATTER.format(
          new Date(lastEventTimestamp),
        )
      : "waiting…";

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLocalNodeIp(value);
    ensureLocalNode();
  };

  const { mode, cursor, range, speed, isPlaying, liveAvailable } = playback;
  const rangeDuration = Math.max(1, range.to - range.from);
  const cursorRatio = Math.min(
    1,
    Math.max(0, (cursor - range.from) / rangeDuration),
  );

  const handleScrub = (value: number) => {
    const nextCursor = range.from + value * rangeDuration;
    onPlaybackChange({ cursor: nextCursor, mode: "historical", isPlaying: false });
  };

  const [historicPanelOpen, setHistoricPanelOpen] = useState(false);
  const [panelRange, setPanelRange] = useState({ from: range.from, to: range.to });
  const [panelError, setPanelError] = useState<string | null>(null);

  useEffect(() => {
    setPanelRange({ from: range.from, to: range.to });
  }, [range.from, range.to]);

  const handleHistoricToggle = () => {
    onPlaybackChange({
      mode: mode === "live" ? "historical" : "live",
      cursor: Date.now(),
    });
    if (mode === "live") {
      setHistoricPanelOpen(true);
    } else {
      setHistoricPanelOpen(false);
      setPanelError(null);
    }
  };

  const handlePanelLoad = async () => {
    if (
      !Number.isFinite(panelRange.from) ||
      !Number.isFinite(panelRange.to) ||
      panelRange.to <= panelRange.from
    ) {
      setPanelError("Start time must be before end time.");
      return;
    }
    setPanelError(null);
    await onHistoricalFetch(panelRange);
    setHistoricPanelOpen(false);
  };

  const startLabel = mounted
    ? new Date(range.from).toLocaleTimeString()
    : "—";
  const endLabel = mounted ? new Date(range.to).toLocaleTimeString() : "—";
  const cursorLabel = mounted ? new Date(cursor).toLocaleTimeString() : "—";

  const statusLabel =
    connectionStatus === "connected"
      ? "Live"
      : connectionStatus === "degraded"
        ? "Degraded"
        : "Connecting";

  return (
    <>
      <nav className="command-nav glass-card">
      <div className="command-brand">
        <Image
          src="/default.svg"
          alt="Monad Flow mark"
          width={88}
          height={88}
          priority
        />
        <div className="command-copy">
          <p className="command-title">Monad Flow</p>
          <span className="command-subtitle">Global Packet Sentinel</span>
        </div>
        <div className="command-meta">
          <div className="status-live-chip">
            <span
              className={`status-live-dot status-${connectionStatus}`}
            />
            <span className="status-live-text">{statusLabel}</span>
          </div>
          <div className="command-badges">
            {mode === "historical" ? (
              <span className="badge" data-variant="info">
                Historical
              </span>
            ) : null}
            {mode === "historical" && liveAvailable ? (
              <span className="badge" data-variant="warning">
                Live Available
              </span>
            ) : null}
          </div>
        </div>
      </div>
      <div className="command-playback">
        <div className="command-playback-head">
          <span className="text-label">Live Monitoring</span>
          <div className="command-playback-actions">
            <button
              type="button"
              className={`status-toggle ${isPlaying ? "active" : ""}`}
              onClick={() => onPlaybackChange({ isPlaying: !isPlaying })}
            >
              {isPlaying ? "Pause" : "Play"}
            </button>
            <button
              type="button"
              className={`status-toggle ghost ${mode === "live" ? "active" : ""}`}
            onClick={handleHistoricToggle}
            >
              {mode === "live" ? "Go Historic" : "Return Live"}
            </button>
            <button
              type="button"
              className={`status-toggle ghost ${processTelemetryVisible ? "active" : ""}`}
              onClick={onToggleProcessTelemetry}
              aria-pressed={processTelemetryVisible}
            >
              {processTelemetryVisible ? "Hide Process Panel" : "Show Process Panel"}
            </button>
          </div>
          <div className="command-playback-meta">
            <span className="text-label">Node Time</span>
            <span className="text-number">{nodeTimeLabel}</span>
          </div>
        </div>
        <div className="command-playback-timeline">
          <span>{startLabel}</span>
          <input
            type="range"
            min={0}
            max={1000}
            value={cursorRatio * 1000}
            onChange={(event) => handleScrub(Number(event.target.value) / 1000)}
          />
          <span>{endLabel}</span>
        </div>
        <div className="command-playback-status">
          <div className="command-playback-current">
            <span className="text-label">Current</span>
            <span className="text-number">{cursorLabel}</span>
          </div>
          <div className="command-playback-speed-inline">
            {SPEED_OPTIONS.map((option) => (
              <button
                key={option}
                type="button"
                className={`status-toggle ${speed === option ? "active" : ""}`}
                onClick={() => onPlaybackChange({ speed: option })}
              >
                {option}×
              </button>
            ))}
          </div>
        </div>
      </div>
      <form className="command-actions" onSubmit={handleSubmit}>
        <label className="text-label" htmlFor="local-node-ip">
          Local Node IP
        </label>
        <input
          id="local-node-ip"
          value={value}
          onChange={(event) => setValue(event.target.value)}
          className="command-input"
          placeholder="e.g. 10.0.1.100"
        />
        <button type="submit" className="status-toggle">
          Apply
        </button>
      </form>
      </nav>
      {historicPanelOpen ? (
        <div className="historical-overlay" role="dialog" aria-modal="true">
          <div className="historical-panel">
            <header className="historical-panel-head">
              <p className="text-title">Historical Playback</p>
              <button
              type="button"
              className="status-toggle ghost"
              onClick={() => setHistoricPanelOpen(false)}
            >
              Close
            </button>
          </header>
          <div className="historical-controls-block">
            <label className="historical-range-label" htmlFor="panel-from">
              From
            </label>
            <input
              id="panel-from"
              type="datetime-local"
              value={formatDatetimeLocal(panelRange.from)}
              onChange={(event) =>
                setPanelRange((prev) => ({
                  ...prev,
                  from: parseDatetimeLocal(event.target.value, prev.from),
                }))
              }
              className="historical-range-input"
            />
            <label className="historical-range-label" htmlFor="panel-to">
              To
            </label>
            <input
              id="panel-to"
              type="datetime-local"
              value={formatDatetimeLocal(panelRange.to)}
              onChange={(event) =>
                setPanelRange((prev) => ({
                  ...prev,
                  to: parseDatetimeLocal(event.target.value, prev.to),
                }))
              }
              className="historical-range-input"
            />
              <div className="historical-panel-actions">
                <button
                  type="button"
                  className={`status-toggle ${historicalLoading ? "active" : ""}`}
                  onClick={handlePanelLoad}
                  disabled={historicalLoading}
                >
                  {historicalLoading ? "Loading…" : "Load Logs"}
                </button>
              </div>
            {panelError ? (
              <p className="historical-error" role="status">
                {panelError}
              </p>
            ) : null}
          </div>
        </div>
      </div>
      ) : null}
    </>
  );
}
