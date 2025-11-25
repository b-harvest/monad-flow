"use client";

import { useEffect } from "react";
import type {
  MonitoringEvent,
  MonitoringSeverity,
  MonadNode,
  NodeTelemetryDigest,
  OutboundRouterEventSummary,
} from "@/types/monad";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";

const createId = () =>
  `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;

const randomBetween = (min: number, max: number) =>
  Math.random() * (max - min) + min;

const severityPalette: MonitoringSeverity[] = ["info", "warning", "critical"];

const BPF_FUNCTIONS = [
  "_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision...",
  "_ZN5monad10StateStore6commitEv",
  "_ZN5monad12ConsensusVM13apply_payload",
];

const SYSTEM_SERVICES = [
  "monad-node.service",
  "monad-bft.service",
  "monad-perf.service",
];

const THREAD_NAMES = ["consensus_thd", "scheduler_thd", "perf_collector"];

const PERF_EVENTS = [
  "cycles",
  "instructions",
  "cache-misses",
  "branch-misses",
];

const ROUTER_STATUS: OutboundRouterEventSummary["status"][] = [
  "delivered",
  "pending",
  "blocked",
];

const ROUTER_MESSAGE_TYPES = [
  { type: 1, label: "PeerDiscovery" },
  { type: 2, label: "FullNodesGroup" },
  { type: 3, label: "AppMessage" },
  { type: 4, label: "StateSync" },
  { type: 5, label: "BlockSync" },
];

function createTelemetrySnapshot(node: MonadNode): NodeTelemetryDigest {
  const now = Date.now();
  const iso = new Date(now).toISOString();
  return {
    bpfTrace: [
      {
        eventType: "enter",
        funcName: BPF_FUNCTIONS[Math.floor(Math.random() * BPF_FUNCTIONS.length)],
        timestamp: iso,
        detail: "caller: main+0x12a4 args[0]=0x7ffe",
      },
      {
        eventType: "exit",
        funcName: "_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision...",
        timestamp: iso,
        durationNs: `${Math.round(1_200_000 + Math.random() * 800_000)}`,
        detail: "return 0x1",
      },
    ],
    systemLogs: [
      {
        service: SYSTEM_SERVICES[Math.floor(Math.random() * SYSTEM_SERVICES.length)],
        timestamp: iso.replace("T", " ").replace("Z", ""),
        message: `${node.name} heartbeat round ${Math.floor(Math.random() * 1000)}`,
      },
    ],
    offCpu: {
      processName: "monad-execution",
      durationUs: Math.round(200 + Math.random() * 4000),
      stack: ["futex_wait_queue", "futex_wait", "__pthread_mutex_lock", "monad::StateDB::commit"],
    },
    scheduler: {
      threadName: THREAD_NAMES[Math.floor(Math.random() * THREAD_NAMES.length)],
      waitDeltaMs: Number((Math.random() * 12).toFixed(2)),
      runDeltaMs: Number((Math.random() * 900 + 50).toFixed(2)),
      ctxSwitches: `${Math.floor(10 + Math.random() * 60)}`,
    },
    perfStat: PERF_EVENTS.map((event) => ({
      event,
      value: `${Math.floor(1_000_000 + Math.random() * 8_000_000)}`,
      metricVal:
        event === "cycles"
          ? `${(2 + Math.random()).toFixed(2)} GHz`
          : event === "instructions"
            ? `${(0.7 + Math.random() * 0.8).toFixed(2)} IPC`
            : undefined,
      runPct: "100.00%",
    })),
    turboStat: {
      core: `${Math.floor(Math.random() * 8)}`,
      avgMhz: Number((2400 + Math.random() * 400).toFixed(1)),
      busyPct: Number((70 + Math.random() * 25).toFixed(1)),
      ipc: Number((1 + Math.random() * 0.8).toFixed(2)),
      pkgWatt: Number((38 + Math.random() * 15).toFixed(2)),
    },
    updatedAt: now,
  };
}

export function useMockPulseStream() {
  useEffect(() => {
    const proposalInterval = setInterval(() => {
      const state = useNodePulseStore.getState();
      const leader = state.nodes.find((node) => node.role === "leader");
      const validators = state.nodes.filter((node) => node.id !== leader?.id);
      if (!leader || validators.length === 0) {
        return;
      }

      const target =
        validators[Math.floor(Math.random() * validators.length)];
      const now = Date.now();
      state.addEffect({
        id: `proposal-${now}`,
        type: "proposal",
        fromNodeId: leader.id,
        toNodeId: target.id,
        createdAt: now,
        ttl: 1600,
      });
      state.pushEvent({
        id: createId(),
        timestamp: now,
        nodeId: leader.id,
        label: `Proposal • Round ${state.metrics.round + 1}`,
        detail: `${leader.name} broadcasting payload to ${target.name}`,
        severity: "info",
      });

      setTimeout(() => {
        const current = useNodePulseStore.getState();
        current.addEffect({
          id: `vote-${Date.now()}`,
          type: "vote",
          fromNodeId: target.id,
          createdAt: Date.now(),
          ttl: 1000,
        });
        const voteEvent: MonitoringEvent = {
          id: createId(),
          timestamp: Date.now(),
          nodeId: target.id,
          label: "Vote Received",
          detail: `${target.name} acknowledged round ${current.metrics.round + 1}`,
          severity:
            severityPalette[
              Math.floor(Math.random() * severityPalette.length)
            ],
        };
        current.pushEvent(voteEvent);
      }, randomBetween(350, 600));
    }, 2600);

    const metricsInterval = setInterval(() => {
      const state = useNodePulseStore.getState();
      state.setMetrics({
        round: state.metrics.round + 1,
        tps: Math.max(
          750,
          Math.min(1800, state.metrics.tps + randomBetween(-120, 120)),
        ),
        blockHeight: state.metrics.blockHeight + 1,
        avgBlockTime: Math.max(0.85, randomBetween(0.95, 1.25)),
        networkHealth: Math.max(
          70,
          Math.min(100, state.metrics.networkHealth + randomBetween(-2, 3)),
        ),
      });
    }, 4300);

    const failureInterval = setInterval(() => {
      if (Math.random() > 0.4) {
        return;
      }
      const state = useNodePulseStore.getState();
      const leader = state.nodes.find((node) => node.role === "leader");
      const validators = state.nodes.filter((node) => node.id !== leader?.id);
      if (!leader || validators.length === 0) {
        return;
      }

      state.setNodeState(leader.id, "failed");
      const failureTimestamp = Date.now();
      state.addEffect({
        id: `failure-${failureTimestamp}`,
        type: "pulse",
        fromNodeId: leader.id,
        createdAt: failureTimestamp,
        ttl: 3200,
      });
      state.pushEvent({
        id: createId(),
        timestamp: failureTimestamp,
        nodeId: leader.id,
        label: "Leader Failure",
        detail: `${leader.name} timed out – recovery initiated`,
        severity: "critical",
      });
      state.setAlert({
        id: createId(),
        title: "Leader Failure Detected",
        description: `Round ${state.metrics.round} stalled on ${leader.name}. Launching recovery.`,
        severity: "critical",
        createdAt: failureTimestamp,
      });

      setTimeout(() => {
        const nextState = useNodePulseStore.getState();
        const newLeader =
          validators[Math.floor(Math.random() * validators.length)];
        nextState.rotateLeader(newLeader.id);
        nextState.setNodeState(leader.id, "active");
        nextState.setNodeState(newLeader.id, "leader");
        const recoveryEvent: MonitoringEvent = {
          id: createId(),
          timestamp: Date.now(),
          nodeId: newLeader.id,
          label: "Leader Switchover",
          detail: `${newLeader.name} promoted – consensus resumed`,
          severity: "info",
        };
        nextState.pushEvent(recoveryEvent);
        nextState.setAlert({
          id: createId(),
          title: "New Leader Elected",
          description: `${newLeader.name} is stabilizing round ${nextState.metrics.round + 1}.`,
          severity: "info",
          createdAt: Date.now(),
        });
      }, 4200);
    }, 22000);

    const cleanupInterval = setInterval(
      () => useNodePulseStore.getState().pruneEffects(),
      900,
    );

    const telemetryInterval = setInterval(() => {
      const state = useNodePulseStore.getState();
      if (state.nodes.length === 0) return;
      const target = state.nodes[Math.floor(Math.random() * state.nodes.length)];
      state.upsertTelemetry(target.id, createTelemetrySnapshot(target));
    }, 4200);

    const routerInterval = setInterval(() => {
      const state = useNodePulseStore.getState();
      const localNode = state.nodes.find((node) => node.isLocal) ?? state.nodes[0];
      if (!localNode) return;
      const messageMeta =
        ROUTER_MESSAGE_TYPES[Math.floor(Math.random() * ROUTER_MESSAGE_TYPES.length)];
      const routerEvent: OutboundRouterEventSummary = {
        id: createId(),
        messageType: messageMeta.type,
        appMessageHash: `0x${cryptoRandomHex(20)}`,
        timestamp: Date.now(),
        peer: `node-${Math.floor(Math.random() * 999)}`,
        size: Math.floor(512 + Math.random() * 4096),
        status: ROUTER_STATUS[Math.floor(Math.random() * ROUTER_STATUS.length)],
        detail: messageMeta.label,
      };
      state.pushRouterEvent(routerEvent);
    }, 3800);

    return () => {
      clearInterval(proposalInterval);
      clearInterval(metricsInterval);
      clearInterval(failureInterval);
      clearInterval(cleanupInterval);
      clearInterval(telemetryInterval);
      clearInterval(routerInterval);
    };
  }, []);
}

function cryptoRandomHex(size: number) {
  let hex = "";
  const chars = "abcdef0123456789";
  for (let i = 0; i < size * 2; i += 1) {
    hex += chars[Math.floor(Math.random() * chars.length)];
  }
  return hex;
}
