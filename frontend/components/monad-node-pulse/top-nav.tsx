"use client";

import Image from "next/image";

interface CommandNavProps {
  streamOpen: boolean;
  onToggleStream: () => void;
  debugOpen: boolean;
  onToggleDebug: () => void;
  socketOpen: boolean;
  onToggleSocket: () => void;
}

export function CommandNav({
  streamOpen,
  onToggleStream,
  debugOpen,
  onToggleDebug,
  socketOpen,
  onToggleSocket,
}: CommandNavProps) {
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
          <span className="command-subtitle">Chain Diagnostics Console</span>
        </div>
      </div>
      <div className="command-actions">
        <button
          type="button"
          className={`status-toggle ${streamOpen ? "active" : ""}`}
          onClick={onToggleStream}
          aria-pressed={streamOpen}
        >
          Socket Stream
        </button>
        <button
          type="button"
          className={`status-toggle ${debugOpen ? "active" : ""}`}
          onClick={onToggleDebug}
          aria-pressed={debugOpen}
        >
          DTO Inspector
        </button>
        <button
          type="button"
          className={`status-toggle ${socketOpen ? "active" : ""}`}
          onClick={onToggleSocket}
          aria-pressed={socketOpen}
        >
          Socket Live
        </button>
      </div>
    </nav>
  );
}
