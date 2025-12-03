"use client";

import { useMemo, useState, useEffect } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import type { ForwardedTxSummary } from "@/lib/monad/slices/router-slice";

export function TransactionFeedPanel() {
  const [hydrated, setHydrated] = useState(false);
  const proposalSnapshots = useNodePulseStore(
    (state) => state.proposalSnapshots,
  );
  const forwardedTxs = useNodePulseStore(
    (state) => state.forwardedTxs,
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

  const forwarded = useMemo<ForwardedTxSummary[]>(() => {
    if (!forwardedTxs || forwardedTxs.length === 0) return [];
    return forwardedTxs.slice(0, 10);
  }, [forwardedTxs]);

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
              <ul className="system-log-messages tx-feed-list">
                {transactions.map((tx, idx) => (
                  <li
                    key={`${tx.hash}-${idx}`}
                    className="system-log-message tx-card"
                  >
                    <div className="tx-card-head">
                      <span className="tx-timestamp">
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
                      <span className="tx-round-pill">
                        Round&nbsp;
                        <span className="tx-round-number">
                          #{tx.proposalRound}
                        </span>
                      </span>
                    </div>

                    <div className="tx-row">
                      <span className="tx-label">Hash</span>
                      <span
                        className="tx-value tx-hash"
                        title={tx.hash}
                      >
                        {formatHash(tx.hash)}
                      </span>
                    </div>

                    {tx.to && (
                      <div className="tx-row">
                        <span className="tx-label">To</span>
                        <span
                          className="tx-value tx-address"
                          title={tx.to}
                        >
                          {tx.to}
                        </span>
                      </div>
                    )}

                    {tx.value && tx.value !== "0x0" && (
                      <div className="tx-row">
                        <span className="tx-label">Value</span>
                        <span className="tx-value tx-value">
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
              <span className="pid-updated">
                {forwarded.length > 0 ? "Live" : "Waiting"}
              </span>
            </div>
            {!hydrated || forwarded.length === 0 ? (
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
                Waiting for ForwardedTxs…
              </div>
            ) : (
              <ul className="system-log-messages tx-feed-list">
                {forwarded.map((tx, idx) => (
                  <li
                    key={`${tx.hash}-${idx}`}
                    className="system-log-message tx-card"
                  >
                    <div className="tx-card-head">
                      <span className="tx-timestamp">
                        {tx.timestamp
                          ? new Date(tx.timestamp).toLocaleTimeString([], {
                              hour12: false,
                              hour: "2-digit",
                              minute: "2-digit",
                              second: "2-digit",
                            })
                          : "—"}
                      </span>
                    </div>
                    <div className="tx-row">
                      <span className="tx-label">Hash</span>
                      <span
                        className="tx-value tx-hash"
                        title={tx.hash}
                      >
                        {formatHash(tx.hash)}
                      </span>
                    </div>
                    {tx.to && (
                      <div className="tx-row">
                        <span className="tx-label">To</span>
                        <span
                          className="tx-value tx-address"
                          title={tx.to}
                        >
                          {tx.to}
                        </span>
                      </div>
                    )}
                    {tx.value && tx.value !== "0x0" && (
                      <div className="tx-row">
                        <span className="tx-label">Value</span>
                        <span className="tx-value tx-value">
                          {tx.value}
                        </span>
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}

function formatHash(hash?: string) {
  if (!hash) return "Unknown Tx";
  if (hash.length <= 18) return hash;
  return `${hash.slice(0, 10)}…${hash.slice(-6)}`;
}
