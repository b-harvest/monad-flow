"use client";

import { useMemo, useState, useEffect } from "react";
import type { SystemLogEvent } from "@/lib/api/system-log";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";

export function SystemLogPanel() {
  const [hydrated, setHydrated] = useState(false);
  const events = useNodePulseStore((state) => state.systemLogEvents);

  useEffect(() => {
    setHydrated(true);
  }, []);

  const groups = useMemo(() => {
    const entries = Object.entries(events);
    if (entries.length === 0) return [];

    return entries
      .sort(([unitA], [unitB]) => unitA.localeCompare(unitB))
      .map(([unit, logs]) => {
        // Logs are already limited to 10 in store, just sort them for display
        const sorted = [...logs].sort(
          (a, b) =>
            (Date.parse(b.timestamp) || 0) - (Date.parse(a.timestamp) || 0),
        );
        const latestTs = sorted[0] ? Date.parse(sorted[0].timestamp) || 0 : 0;
        return {
          unit,
          latestTs,
          messages: sorted,
        };
      });
  }, [events]);

  return (
    <section className="system-log-panel hud-panel">
      <header className="panel-header">
        <div>
          <span className="text-label">System Log</span>
        </div>
      </header>
      <div className="system-log-strip">
        {!hydrated || groups.length === 0 ? (
          <div className="system-log-card">
            <span className="text-label">System Log</span>
            <p className="pid-placeholder">Waiting for system logs…</p>
          </div>
        ) : (
          groups.map((group) => (
            <div key={group.unit} className="system-log-card">
              <div className="system-log-card-head">
                <div>
                  <span className="text-label">Unit</span>
                  <p className="text-title">{group.unit}</p>
                </div>
                <span className="pid-updated">
                  {group.latestTs ? formatRelative(group.latestTs) : "—"}
                </span>
              </div>
              <ul className="system-log-messages">
                {group.messages.map((message) => (
                  <li key={message._id} className="system-log-message">
                    <span className="system-log-ts">
                      {new Date(message.timestamp).toLocaleTimeString()}
                    </span>
                    <p>{message.message}</p>
                  </li>
                ))}
              </ul>
            </div>
          ))
        )}
      </div>
    </section>
  );
}

function formatRelative(timestamp: number) {
  if (!timestamp) return "—";
  const diff = Date.now() - timestamp;
  const seconds = Math.round(diff / 1000);
  if (seconds < 1) return "just now";
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.round(minutes / 60);
  return `${hours}h ago`;
}
