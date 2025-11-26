"use client";

import { useMemo, useState, useEffect, useSyncExternalStore } from "react";
import type { SystemLogEvent } from "@/lib/api/system-log";
import {
  getSystemLogEvents,
  subscribeToSystemLogEvents,
} from "@/lib/storage/system-log-cache";

export function SystemLogPanel() {
  const [hydrated, setHydrated] = useState(false);
  const events = useSyncExternalStore(
    subscribeToSystemLogEvents,
    getSystemLogEvents,
    getSystemLogEvents,
  );

  useEffect(() => {
    setHydrated(true);
  }, []);

  const groups = useMemo(() => {
    if (events.length === 0) return [];
    const grouped = new Map<string, SystemLogEvent[]>();
    events.slice(-100).forEach((event) => {
      const bucket = grouped.get(event.unit) ?? [];
      bucket.push(event);
      grouped.set(event.unit, bucket);
    });
    return Array.from(grouped.entries())
      .sort(([unitA], [unitB]) => unitA.localeCompare(unitB))
      .map(([unit, logs]) => {
        const sorted = [...logs].sort(
          (a, b) =>
            (Date.parse(a.timestamp) || 0) - (Date.parse(b.timestamp) || 0),
        );
        const recent = sorted.slice(-4).reverse();
        const latestTs = recent[0] ? Date.parse(recent[0].timestamp) || 0 : 0;
        return {
          unit,
          latestTs,
          messages: recent,
        };
      })
      .slice(0, 2);
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
