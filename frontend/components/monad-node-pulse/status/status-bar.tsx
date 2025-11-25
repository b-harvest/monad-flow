"use client";

import type { PlaybackState } from "@/types/monad";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { useNow } from "@/lib/hooks/use-now";

interface StatusBarProps {
  connectionStatus: "connected" | "degraded" | "lost";
  lastEventTimestamp: number;
  playbackMode: PlaybackState["mode"];
  liveAvailable: boolean;
}

const UTC_FORMATTER = new Intl.DateTimeFormat("en-GB", {
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hour12: false,
  timeZone: "UTC",
});

export function StatusBar({
  lastEventTimestamp,
  playbackMode,
  liveAvailable,
}: StatusBarProps) {
  const preferences = useNodePulseStore((state) => state.preferences);
  const setPreferences = useNodePulseStore((state) => state.setPreferences);
  const now = useNow(1000);
  const nowUtc =
    now === null ? "—" : UTC_FORMATTER.format(new Date(now));

  const lastEventDiff =
    now === null
      ? null
      : Math.max(0, Math.round((now - lastEventTimestamp) / 1000));

  return (
    <header className="status-bar glass-card">
      <div className="status-cluster">
        <div className="status-brand">
          <div className="status-brand-copy">
            <span className="text-label text-secondary">Monad Flow</span>
            <span className="status-brand-subtitle">BHarvest Monad Tooling</span>
          </div>
          <div className="status-live-chip">
            <span className="status-live-dot" />
            <span className="status-live-text">Live</span>
          </div>
        </div>
        <div className="status-meta">
          {playbackMode === "historical" ? (
            <span className="badge" data-variant="info">
              Historical
            </span>
          ) : null}
          {playbackMode === "historical" && liveAvailable ? (
            <span className="badge" data-variant="warning">
              Live Available
            </span>
          ) : null}
        </div>
      </div>

      <div className="status-clock">
        <div>
          <span className="text-label">UTC</span>
          <span className="text-number">{nowUtc}</span>
        </div>
        <div>
          <span className="text-label">Last Event</span>
          <span className="text-number">
            {lastEventDiff === null
              ? "Syncing…"
              : lastEventDiff === 0
                ? "Just now"
                : `${lastEventDiff}s ago`}
          </span>
        </div>
      </div>

      <div className="status-controls">
        <button
          type="button"
          className={`status-toggle status-toggle-vertical ${preferences.autoRotate ? "active" : ""}`}
          onClick={() =>
            setPreferences({ autoRotate: !preferences.autoRotate })
          }
          aria-pressed={preferences.autoRotate}
        >
          Auto-Rotate
        </button>
        <button
          type="button"
          className={`status-toggle status-toggle-vertical ${preferences.showParticles ? "active" : ""}`}
          onClick={() =>
            setPreferences({ showParticles: !preferences.showParticles })
          }
          aria-pressed={preferences.showParticles}
        >
          Particles
        </button>
      </div>
    </header>
  );
}
