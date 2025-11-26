"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import type { PlaybackState } from "@/types/monad";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { useNow } from "@/lib/hooks/use-now";

interface CommandNavProps {
  connectionStatus: "connected" | "degraded" | "lost";
  playback: PlaybackState;
  onPlaybackChange: (patch: Partial<PlaybackState>) => void;
  processTelemetryVisible: boolean;
  onToggleProcessTelemetry: () => void;
}

const UTC_FORMATTER = new Intl.DateTimeFormat("en-GB", {
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hour12: false,
  timeZone: "UTC",
});

const SPEED_OPTIONS: PlaybackState["speed"][] = [0.25, 0.5, 1, 2, 4];

export function CommandNav({
  connectionStatus,
  playback,
  onPlaybackChange,
  processTelemetryVisible,
  onToggleProcessTelemetry,
}: CommandNavProps) {
  const localIp = useNodePulseStore((state) => state.localNodeIp);
  const setLocalNodeIp = useNodePulseStore((state) => state.setLocalNodeIp);
  const ensureLocalNode = useNodePulseStore((state) => state.ensureLocalNode);
  const [value, setValue] = useState(localIp);
  const [mounted, setMounted] = useState(false);
  const now = useNow(1000);

  useEffect(() => {
    setValue(localIp);
  }, [localIp]);

  useEffect(() => {
    setMounted(true);
  }, []);

  const nowUtc = now === null ? "—" : UTC_FORMATTER.format(new Date(now));

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

  const startLabel = mounted
    ? new Date(range.from).toLocaleTimeString()
    : "—";
  const endLabel = mounted ? new Date(range.to).toLocaleTimeString() : "—";

  const statusLabel =
    connectionStatus === "connected"
      ? "Live"
      : connectionStatus === "degraded"
        ? "Degraded"
        : "Connecting";

  return (
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
              onClick={() =>
                onPlaybackChange({
                  mode: mode === "live" ? "historical" : "live",
                  cursor: Date.now(),
                })
              }
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
            <span className="text-label">UTC</span>
            <span className="text-number">{nowUtc}</span>
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
        <div className="command-playback-speed">
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
  );
}
