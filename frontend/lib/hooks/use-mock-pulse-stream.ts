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

const ENABLE_MOCK =
  typeof process !== "undefined" &&
  process.env.NEXT_PUBLIC_USE_MOCK === "true";

export function useMockPulseStream() {
  useEffect(() => {
    if (!ENABLE_MOCK) {
      return undefined;
    }
    const state = useNodePulseStore.getState();
    state.setConnectionStatus("connected");
    const localId = state.ensureLocalNode();

    const chunkInterval = setInterval(() => {
      const current = useNodePulseStore.getState();
      const ip = `${Math.floor(Math.random() * 200) + 20}.${Math.floor(Math.random() * 200)}.${Math.floor(Math.random() * 200)}.${Math.floor(Math.random() * 200)}`;
      const port = 3000 + Math.floor(Math.random() * 2000);
      const peerId = current.upsertChunkPeer(ip, port);
      current.addEffect({
        id: `mock-chunk-${Date.now()}`,
        type: "chunk",
        fromNodeId: peerId,
        toNodeId: localId,
        createdAt: Date.now(),
        ttl: 1500,
      });
      const severity =
        severityPalette[Math.floor(Math.random() * severityPalette.length)];
      current.pushEvent({
        id: createId(),
        timestamp: Date.now(),
        nodeId: peerId,
        label: "Chunk Received",
        detail: `${ip}:${port} relayed packet to Monad Validator`,
        severity,
      });
    }, 2600);

    const metricsInterval = setInterval(() => {
      const current = useNodePulseStore.getState();
      current.setMetrics({
        round: current.metrics.round + 1,
        tps: Math.max(
          750,
          Math.min(1800, current.metrics.tps + randomBetween(-120, 120)),
        ),
        blockHeight: current.metrics.blockHeight + 1,
        avgBlockTime: Math.max(0.85, randomBetween(0.95, 1.25)),
        networkHealth: Math.max(
          70,
          Math.min(100, current.metrics.networkHealth + randomBetween(-2, 3)),
        ),
      });
    }, 4300);

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
      clearInterval(chunkInterval);
      clearInterval(metricsInterval);
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
