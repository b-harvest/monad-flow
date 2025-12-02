"use client";

import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { KNOWN_PEERS } from "@/lib/monad/known-peers";
import type { LeaderEvent } from "@/lib/api/leader";

export function LeaderSchedulePanel() {
  const leaders = useNodePulseStore((state) => state.leaders);

  return (
    <section className="hud-panel leader-schedule-panel">
      <header className="panel-header">
        <div>
          <span className="text-label">Leader Schedule</span>
        </div>
      </header>
      {leaders.length === 0 ? (
        <div className="leader-schedule-empty">
          <span className="text-label">Leader Schedule</span>
          <p className="pid-placeholder">Waiting for schedule…</p>
        </div>
      ) : (
        <div className="leader-schedule-body">
          <ul className="leader-schedule-list">
            {[...leaders].reverse().map((leader, index) => (
              <li
                key={leader._id}
                className={`leader-schedule-item ${
                  index === 0 ? "leader-schedule-item-active" : ""
                }`}
              >
                <div className="leader-round">
                  <span className="leader-round-label">Round</span>
                  <span className="leader-round-value">
                    {leader.round.toLocaleString()}
                  </span>
                </div>
                <div className="leader-node">
                  <span className="leader-node-label">Leader</span>
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
  const known = KNOWN_PEERS[rawId] as { name?: string } | undefined;
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
