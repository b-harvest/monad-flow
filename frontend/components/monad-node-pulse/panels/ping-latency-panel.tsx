"use client";

import { useMemo, useState, useEffect } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";

export function PingLatencyPanel() {
  const [hydrated, setHydrated] = useState(false);
  const pings = useNodePulseStore((state) => state.recentPings);

  useEffect(() => {
    setHydrated(true);
  }, []);

  return (
    <section className="ping-latency-panel hud-panel">
      <header className="panel-header">
        <div>
          <span className="text-label">Ping Latency (Latest)</span>
        </div>
      </header>
      <div className="system-log-strip">
        {!hydrated || pings.length === 0 ? (
          <div className="system-log-card">
            <span className="text-label">Ping Latency</span>
            <p className="pid-placeholder">Waiting for pings…</p>
          </div>
        ) : (
          <div className="system-log-card ping-grid">
            {Object.values(
              pings.reduce(
                (acc, ping) => {
                  if (
                    !acc[ping.ip] ||
                    ping.timestamp > acc[ping.ip].timestamp
                  ) {
                    acc[ping.ip] = ping;
                  }
                  return acc;
                },
                {} as Record<string, (typeof pings)[0]>,
              ),
            )
              .sort((a, b) => a.ip.localeCompare(b.ip))
              .map((ping) => (
                <div
                  key={ping.ip}
                  className={`ping-card ${
                    ping.rtt_ms < 80
                      ? "ping-good"
                      : ping.rtt_ms < 150
                        ? "ping-warn"
                        : "ping-bad"
                  }`}
                >
                  <span
                    className="ping-latency-value"
                  >
                    {ping.rtt_ms}ms
                  </span>
                  <div className="ping-node-name">
                    {ping.name !== `${ping.ip}:0` && ping.name !== ping.ip
                      ? ping.name
                      : ping.ip}
                  </div>
                  {ping.name !== `${ping.ip}:0` && ping.name !== ping.ip && (
                    <div className="ping-node-ip">
                      {ping.ip}
                    </div>
                  )}
                </div>
              ))}
          </div>
        )}
      </div>
    </section>
  );
}
