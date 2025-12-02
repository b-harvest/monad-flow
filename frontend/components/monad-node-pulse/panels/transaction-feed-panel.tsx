"use client";

import { useMemo, useState, useEffect } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";

export function TransactionFeedPanel() {
  const [hydrated, setHydrated] = useState(false);
  const proposalSnapshots = useNodePulseStore(
    (state) => state.proposalSnapshots,
  );

  useEffect(() => {
    setHydrated(true);
  }, []);

  const transactions = useMemo(() => {
    if (!proposalSnapshots || proposalSnapshots.length === 0) return [];

    // Flatten transactions from all snapshots and take the latest 50
    const allTxs = proposalSnapshots.flatMap((snapshot) => {
      return (snapshot.transactions || []).map((tx) => ({
        ...tx,
        proposalRound: snapshot.round,
        proposalTimestamp: snapshot.timestampNs / 1_000_000, // Convert ns to ms
      }));
    });

    // Sort by timestamp descending (assuming proposal order is roughly chronological)
    // Since we flatMap from snapshots which are likely ordered, we can just reverse
    return allTxs.reverse().slice(0, 10);
  }, [proposalSnapshots]);

  return (
    <section className="transaction-feed-panel hud-panel">
      <header className="panel-header">
        <div>
          <span className="text-label">Transaction Feed</span>
        </div>
      </header>
      <div className="system-log-strip" style={{ display: "flex", gap: "16px" }}>
        {/* Left Column: Proposal Transactions */}
        <div style={{ flex: 1, minWidth: 0 }}>
          {!hydrated || transactions.length === 0 ? (
            <div className="system-log-card">
              <span className="text-label">Proposal Transactions</span>
              <p className="pid-placeholder">Waiting for transactions…</p>
            </div>
          ) : (
            <div className="system-log-card">
              <div className="system-log-card-head">
                <div>
                  <span className="text-label">Proposal Transactions</span>
                </div>
                <span className="pid-updated">Live</span>
              </div>
              <ul
                className="system-log-messages"
                style={{ minHeight: "600px", maxHeight: "800px", overflowY: "auto" }}
              >
                {transactions.map((tx, idx) => (
                  <li
                    key={`${tx.hash}-${idx}`}
                    className="system-log-message"
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      gap: "4px",
                      alignItems: "flex-start",
                      padding: "12px",
                      borderBottom: "1px solid rgba(255, 255, 255, 0.05)",
                    }}
                  >
                    <span
                      className="system-log-ts"
                      style={{ fontSize: "0.75rem", opacity: 0.5, marginBottom: "2px" }}
                    >
                      {tx.proposalTimestamp
                        ? new Date(tx.proposalTimestamp).toLocaleTimeString([], {
                            hour12: false,
                            hour: "2-digit",
                            minute: "2-digit",
                            second: "2-digit",
                            fractionalSecondDigits: 3,
                          })
                        : "—"}
                    </span>

                    <div className="flex items-center gap-2 w-full">
                      <span className="text-white/40 text-[10px] uppercase tracking-wider w-10 shrink-0">
                        Hash
                      </span>
                      <span
                        className="text-white font-mono text-xs truncate"
                        style={{ color: "#a5b4fc" }}
                      >
                        {tx.hash || "Unknown Tx"}
                      </span>
                    </div>

                    <div className="flex items-center gap-2 w-full">
                      <span className="text-white/40 text-[10px] uppercase tracking-wider w-10 shrink-0">
                        Round
                      </span>
                      <span className="text-white font-mono text-xs">
                        #{tx.proposalRound}
                      </span>
                    </div>

                    {tx.to && (
                      <div className="flex items-center gap-2 w-full">
                        <span className="text-white/40 text-[10px] uppercase tracking-wider w-10 shrink-0">
                          To
                        </span>
                        <span
                          className="font-mono text-xs truncate"
                          style={{ color: "#86efac" }}
                        >
                          {tx.to}
                        </span>
                      </div>
                    )}

                    {tx.value && tx.value !== "0x0" && (
                      <div className="flex items-center gap-2 w-full">
                        <span className="text-white/40 text-[10px] uppercase tracking-wider w-10 shrink-0">
                          Val
                        </span>
                        <span className="text-white/80 font-mono text-xs truncate">
                          {tx.value}
                        </span>
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {/* Right Column: Forwarded Transactions (Placeholder) */}
        <div style={{ flex: 1, minWidth: 0 }}>
          <div className="system-log-card" style={{ height: "100%" }}>
            <div className="system-log-card-head">
              <div>
                <span className="text-label">Forwarded Transactions</span>
              </div>
              <span className="pid-updated">Coming Soon</span>
            </div>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                height: "100px",
                color: "rgba(255, 255, 255, 0.4)",
                fontSize: "0.9em",
                fontStyle: "italic",
              }}
            >
              Preparing ForwardedTx Stream...
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
