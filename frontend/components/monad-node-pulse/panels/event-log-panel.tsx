"use client";

import type { MonitoringEvent } from "@/types/monad";

interface EventLogPanelProps {
  events: MonitoringEvent[];
}

const formatter = new Intl.DateTimeFormat("en-GB", {
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hour12: false,
});

const SEVERITY_VARIANT: Record<
  MonitoringEvent["severity"],
  "info" | "warning" | "danger"
> = {
  info: "info",
  warning: "warning",
  critical: "danger",
};

export function EventLogPanel({ events }: EventLogPanelProps) {
  const reversed = [...events]
    .sort((a, b) => b.timestamp - a.timestamp)
    .slice(0, 50);
  return (
    <section className="hud-panel event-log-panel">
      <header className="panel-header">
        <div>
          <span className="text-label">Event Feed</span>
          <p className="text-title">Socket.IO Stream</p>
        </div>
        <span className="badge" data-variant="info">
          {reversed.length} / 50
        </span>
      </header>
      <ol className="event-log-list">
        {reversed.map((event) => (
          <li key={event.id} className="event-log-item">
            <div>
              <span className="event-time">
                {formatter.format(new Date(event.timestamp))}
              </span>
              <p className="event-label">{event.label}</p>
              <p className="event-detail">{event.detail}</p>
            </div>
            <span className="badge" data-variant={SEVERITY_VARIANT[event.severity]}>
              {event.severity}
            </span>
          </li>
        ))}
      </ol>
    </section>
  );
}
