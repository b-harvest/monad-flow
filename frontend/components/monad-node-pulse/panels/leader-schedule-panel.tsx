"use client";

import { useEffect, useMemo, useState } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { KNOWN_PEERS } from "@/lib/monad/known-peers";
import type { LeaderEvent } from "@/lib/api/leader";
import { fetchLeaderSchedule } from "@/lib/api/leader";

const FUTURE_RANGE = 5;

export function LeaderSchedulePanel() {
  const proposalSnapshots = useNodePulseStore(
    (state) => state.proposalSnapshots,
  );

  const latestProposal = useMemo(() => {
    const count = proposalSnapshots.length;
    return count > 0 ? proposalSnapshots[count - 1] : null;
  }, [proposalSnapshots]);

  const [baseRound, setBaseRound] = useState<number | null>(
    null,
  );
  const [leaders, setLeaders] = useState<LeaderEvent[]>(
    [],
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!latestProposal) {
      return;
    }
    const nextBase = latestProposal.round;
    if (baseRound !== null && baseRound === nextBase) {
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);

    fetchLeaderSchedule(nextBase, FUTURE_RANGE)
      .then((events) => {
        if (cancelled) return;
        setBaseRound(nextBase);
        setLeaders(events);
      })
      .catch((fetchError: unknown) => {
        if (cancelled) return;
        setError(
          fetchError instanceof Error
            ? fetchError.message
            : "Failed to load leader schedule",
        );
        setLeaders([]);
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [latestProposal, baseRound]);

  const futureLeaders = useMemo(() => {
    if (!latestProposal) {
      return [];
    }
    return leaders
      .filter((leader) => leader.round > latestProposal.round)
      .slice(0, FUTURE_RANGE);
  }, [leaders, latestProposal]);

  const hasLeaders = futureLeaders.length > 0;

  return (
    <section className="hud-panel leader-schedule-panel">
      <header className="panel-header">
        <div>
          <span className="text-label">Leader Schedule</span>
        </div>
      </header>
      {!hasLeaders ? (
        <div className="leader-schedule-empty">
          <span className="text-label">Leader Schedule</span>
          <p className="pid-placeholder">
            {loading
              ? "Loading next leaders…"
              : error ?? "Waiting for proposal…"}
          </p>
        </div>
      ) : (
        <div className="leader-schedule-body">
          <ul className="leader-schedule-list">
            {futureLeaders.map((leader, index) => (
              <li
                key={leader._id}
                className={`leader-schedule-item ${
                  index === 0 ? "leader-schedule-item-active" : ""
                }`}
              >
                <div className="leader-round">
                  <span className="leader-round-label">
                    Round
                  </span>
                  <span className="leader-round-value">
                    {leader.round.toLocaleString()}
                  </span>
                </div>
                <div className="leader-node">
                  <span className="leader-node-label">
                    Leader
                  </span>
                  <span
                    className="leader-node-id"
                    title={leader.node_id}
                  >
                    {getLeaderDisplayName(leader)}
                  </span>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  );
}

function getLeaderDisplayName(leader: LeaderEvent): string {
  const rawId = leader.node_id.startsWith("0x")
    ? leader.node_id.slice(2)
    : leader.node_id;
  const known = KNOWN_PEERS[rawId] as
    | { name?: string }
    | undefined;
  if (known && typeof known.name === "string" && known.name.length > 0) {
    return known.name;
  }
  return formatNodeId(leader.node_id);
}

function formatNodeId(nodeId: string) {
  if (!nodeId) return "—";
  if (nodeId.length <= 18) return nodeId;
  return `${nodeId.slice(0, 10)}…${nodeId.slice(-6)}`;
}
