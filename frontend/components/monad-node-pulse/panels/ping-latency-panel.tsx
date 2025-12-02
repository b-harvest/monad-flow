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
          <div
            className="system-log-card"
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(8, 1fr)",
              gap: "8px",
              padding: "8px",
            }}
          >
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
                  className="system-log-message"
                  style={{
                    display: "flex",
                    flexDirection: "column",
                    alignItems: "center",
                    justifyContent: "center",
                    background: "rgba(255, 255, 255, 0.05)",
                    borderRadius: "4px",
                    padding: "8px",
                    textAlign: "center",
                    minHeight: "100px",
                  }}
                >
                  <span
                    className="system-log-ts"
                    style={{
                      fontSize: "1.2em",
                      fontWeight: "bold",
                      color: ping.rtt_ms < 100 ? "#4ade80" : "#fbbf24",
                      marginBottom: "4px",
                    }}
                  >
                    {ping.rtt_ms}ms
                  </span>
                  <div className="text-white font-medium text-xs truncate w-full">
                    {ping.name !== `${ping.ip}:0` && ping.name !== ping.ip
                      ? ping.name
                      : ping.ip}
                  </div>
                  {ping.name !== `${ping.ip}:0` && ping.name !== ping.ip && (
                    <div className="text-[10px] text-white/40 truncate w-full">
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
