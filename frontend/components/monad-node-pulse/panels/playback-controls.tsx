"use client";

import type { ChangeEvent } from "react";
import type { PlaybackState } from "@/types/monad";

interface PlaybackControlsProps {
  playback: PlaybackState;
  onChange: (patch: Partial<PlaybackState>) => void;
}

const SPEED_OPTIONS: PlaybackState["speed"][] = [0.25, 0.5, 1, 2, 4];

export function PlaybackControls({
  playback,
  onChange,
}: PlaybackControlsProps) {
  const { mode, cursor, range, speed, isPlaying } = playback;
  const rangeDuration = range.to - range.from;
  const cursorRatio =
    rangeDuration === 0 ? 0 : (cursor - range.from) / rangeDuration;

  const handleScrub = (nextRatio: number) => {
    const nextCursor = range.from + nextRatio * rangeDuration;
    onChange({ cursor: nextCursor, mode: "historical", isPlaying: false });
  };

  return (
    <section className="hud-panel playback-panel">
      <header className="panel-header">
        <div>
          <span className="text-label">Historical Forensics</span>
          <p className="text-title">
            {mode === "live" ? "Live Monitoring" : "Time Scrubber"}
          </p>
        </div>
        <button
          type="button"
          className={`status-toggle ${isPlaying ? "active" : ""}`}
          onClick={() => onChange({ isPlaying: !isPlaying })}
        >
          {isPlaying ? "Pause" : "Play"}
        </button>
      </header>

      <div className="timeline">
        <span>{new Date(range.from).toLocaleTimeString()}</span>
        <input
          type="range"
          min={0}
          max={1000}
          value={cursorRatio * 1000}
          onChange={(event: ChangeEvent<HTMLInputElement>) =>
            handleScrub(Number(event.target.value) / 1000)
          }
        />
        <span>{new Date(range.to).toLocaleTimeString()}</span>
      </div>

      <div className="speed-controls">
        {SPEED_OPTIONS.map((option) => (
          <button
            key={option}
            type="button"
            className={`status-toggle ${speed === option ? "active" : ""}`}
            onClick={() => onChange({ speed: option })}
          >
            {option}×
          </button>
        ))}
        <button
          type="button"
          className={`status-toggle ghost ${mode === "live" ? "active" : ""}`}
          onClick={() =>
            onChange({
              mode: mode === "live" ? "historical" : "live",
              cursor: Date.now(),
            })
          }
        >
          {mode === "live" ? "Go Historic" : "Return Live"}
        </button>
      </div>
    </section>
  );
}
