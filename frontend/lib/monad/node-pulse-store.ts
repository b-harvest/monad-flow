import { create } from "zustand";
import type {
  AlertToast,
  ConsensusMetrics,
  MonadNode,
  MonadNodeState,
  MonitoringEvent,
  PlaybackState,
  PulseVisualEffect,
  NodeTelemetryDigest,
  OutboundRouterEventSummary,
  ChunkPacketRecord,
} from "@/types/monad";
import { approximateGeoFromIp } from "@/lib/geo/ip-to-geo";

interface NodePulseState {
  nodes: MonadNode[];
  localNodeIp: string;
  metrics: ConsensusMetrics;
  eventLog: MonitoringEvent[];
  visualEffects: PulseVisualEffect[];
  playback: PlaybackState;
  alert: AlertToast | null;
  lastEventTimestamp: number;
  selectedNodeId: string | null;
  nodeTelemetry: Record<string, NodeTelemetryDigest>;
  routerEvents: OutboundRouterEventSummary[];
  selectedRouterEventId: string | null;
  detailDrawerOpen: boolean;
  chunkQueue: Record<string, ChunkPacketRecord[]>;
  activeAssembly: {
    hash: string;
    packets: ChunkPacketRecord[];
    startedAt: number;
  } | null;
  setNodeState: (nodeId: string, state: MonadNodeState) => void;
  rotateLeader: (nodeId: string) => void;
  pushEvent: (event: MonitoringEvent) => void;
  addEffect: (effect: PulseVisualEffect) => void;
  pruneEffects: (now?: number) => void;
  setMetrics: (metrics: Partial<ConsensusMetrics>) => void;
  patchNode: (nodeId: string, patch: Partial<MonadNode>) => void;
  setPlayback: (patch: Partial<PlaybackState>) => void;
  setAlert: (toast: AlertToast | null) => void;
  setSelectedNode: (nodeId: string | null) => void;
  upsertTelemetry: (nodeId: string, payload: NodeTelemetryDigest) => void;
  ensureLocalNode: () => string;
  upsertChunkPeer: (ip: string, port: number) => string;
  resetNetworkGraph: () => void;
  setConnectionStatus: (
    status: ConsensusMetrics["connectionStatus"],
  ) => void;
  pruneInactiveNodes: (thresholdMs?: number) => void;
  pushChunkPacket: (packet: ChunkPacketRecord) => void;
  triggerChunkAssembly: (hash?: string) => void;
  clearActiveAssembly: () => void;
  pushRouterEvent: (event: OutboundRouterEventSummary) => void;
  setSelectedRouterEvent: (id: string) => void;
  setDetailDrawerOpen: (open: boolean) => void;
}

const LOCAL_NODE_ID = "monad-local-node";
const CHUNK_RING_HEIGHT = -0.5;

const initialMetrics: ConsensusMetrics = {
  epoch: 128,
  round: 512,
  leaderId: LOCAL_NODE_ID,
  tps: 1324,
  blockHeight: 983_455,
  avgBlockTime: 1.1,
  networkHealth: 94,
  connectionStatus: "lost",
  timestamp: Date.now(),
};

const initialPlayback: PlaybackState = {
  mode: "live",
  range: {
    from: Date.now() - 5 * 60 * 1000,
    to: Date.now(),
  },
  cursor: Date.now(),
  isPlaying: false,
  speed: 1,
  liveAvailable: true,
};

const createLocalNode = (ip: string): MonadNode => {
  const geo = approximateGeoFromIp(ip);
  return {
    id: LOCAL_NODE_ID,
    name: "Monad Validator",
    role: "leader",
    ip,
    uptimePct: 100,
    participationRate: 100,
    lastActivity: new Date().toISOString(),
    state: "leader",
    position: [geo.longitude, 0, 0],
    latitude: geo.latitude,
    longitude: geo.longitude,
    isLocal: true,
  };
};

const createChunkNode = (ip: string, port: number): MonadNode => {
  const geo = approximateGeoFromIp(ip, port);
  return {
    id: `chunk-${ip}:${port}`,
    name: `${ip}:${port}`,
    role: "validator",
    ip,
    uptimePct: 92 + Math.random() * 6,
    participationRate: 85 + Math.random() * 10,
    lastActivity: new Date().toISOString(),
    state: "active",
    position: [geo.longitude, CHUNK_RING_HEIGHT, 0],
    latitude: geo.latitude,
    longitude: geo.longitude,
  };
};

export const useNodePulseStore = create<NodePulseState>()((set, get) => ({
  nodes: [],
  localNodeIp: "64.31.48.109",
  metrics: initialMetrics,
  eventLog: [],
  visualEffects: [],
  playback: initialPlayback,
  alert: null,
  lastEventTimestamp: Date.now(),
  selectedNodeId: null,
  nodeTelemetry: {},
  routerEvents: [],
  selectedRouterEventId: null,
  detailDrawerOpen: false,
  chunkQueue: {},
  activeAssembly: null,

  setNodeState: (nodeId, state) =>
    set((stateObj) => ({
      nodes: stateObj.nodes.map((node) =>
        node.id === nodeId ? { ...node, state } : node,
      ),
    })),

  rotateLeader: (nodeId) =>
    set((stateObj) => ({
      nodes: stateObj.nodes.map((node) => {
        if (node.id === nodeId) {
          return {
            ...node,
            role: "leader",
            state: "leader",
          };
        }
        if (node.id === stateObj.metrics.leaderId) {
          return {
            ...node,
            role: "validator",
            state: node.state === "leader" ? "active" : node.state,
          };
        }
        return node;
      }),
      metrics: {
        ...stateObj.metrics,
        leaderId: nodeId,
      },
    })),

  pushEvent: (event) =>
    set((stateObj) => {
      const trimmed = [...stateObj.eventLog, event]
        .sort((a, b) => a.timestamp - b.timestamp)
        .slice(-50);
      return {
        eventLog: trimmed,
        lastEventTimestamp: event.timestamp,
      };
    }),

  addEffect: (effect) =>
    set((stateObj) => ({
      visualEffects: [...stateObj.visualEffects, effect],
    })),

  pruneEffects: (now = Date.now()) =>
    set((stateObj) => ({
      visualEffects: stateObj.visualEffects.filter(
        (effect) => effect.createdAt + effect.ttl > now,
      ),
    })),

  setMetrics: (patch) =>
    set((stateObj) => ({
      metrics: { ...stateObj.metrics, ...patch, timestamp: Date.now() },
    })),

  patchNode: (nodeId, patch) =>
    set((stateObj) => ({
      nodes: stateObj.nodes.map((node) =>
        node.id === nodeId ? { ...node, ...patch } : node,
      ),
    })),

  setPlayback: (patch) =>
    set((stateObj) => ({
      playback: { ...stateObj.playback, ...patch },
    })),

  setAlert: (toast) => set({ alert: toast }),

  setSelectedNode: (nodeId) =>
    set((stateObj) => {
      if (nodeId === null) {
        return { selectedNodeId: null };
      }
      const isLocal = stateObj.nodes.find(
        (node) => node.id === nodeId && node.isLocal,
      );
      return {
        selectedNodeId: isLocal ? nodeId : stateObj.selectedNodeId,
      };
    }),

  upsertTelemetry: (nodeId, payload) =>
    set((stateObj) => ({
      nodeTelemetry: {
        ...stateObj.nodeTelemetry,
        [nodeId]: payload,
      },
    })),

  ensureLocalNode: () => {
    const stateObj = get();
    const existing = stateObj.nodes.find((node) => node.isLocal);
    if (existing) {
      return existing.id;
    }
    const localNode = createLocalNode(stateObj.localNodeIp);
    set({
      nodes: [localNode, ...stateObj.nodes],
      metrics: { ...stateObj.metrics, leaderId: localNode.id },
    });
    return localNode.id;
  },

  upsertChunkPeer: (ip, port) => {
    const nodeId = `chunk-${ip}:${port}`;
    const stateObj = get();
    const existing = stateObj.nodes.find((node) => node.id === nodeId);
    if (existing) {
      set({
        nodes: stateObj.nodes.map((node) =>
          node.id === nodeId
            ? {
                ...node,
                lastActivity: new Date().toISOString(),
                state: "active",
              }
            : node,
        ),
      });
      return nodeId;
    }
    const newNode = createChunkNode(ip, port);
    set({ nodes: [...stateObj.nodes, newNode] });
    return nodeId;
  },

  resetNetworkGraph: () =>
    set(() => ({
      nodes: [],
      selectedNodeId: null,
      chunkQueue: {},
      activeAssembly: null,
    })),

  setConnectionStatus: (status) =>
    set((stateObj) => ({
      metrics: { ...stateObj.metrics, connectionStatus: status },
    })),

  pruneInactiveNodes: (thresholdMs = 60_000) =>
    set((stateObj) => {
      const cutoff = Date.now() - thresholdMs;
      const nodes = stateObj.nodes.filter((node) => {
        if (node.isLocal) {
          return true;
        }
        const last = node.lastActivity
          ? new Date(node.lastActivity).getTime()
          : 0;
        return last >= cutoff;
      });
      const selectedNodeId = nodes.some(
        (node) => node.id === stateObj.selectedNodeId,
      )
        ? stateObj.selectedNodeId
        : null;
      if (nodes.length === stateObj.nodes.length) {
        return stateObj;
      }
      return { nodes, selectedNodeId };
    }),

  pushChunkPacket: (packet) =>
    set((stateObj) => {
      const key = packet.appMessageHash ?? "unmapped";
      const existing = stateObj.chunkQueue[key] ?? [];
      const nextQueue = {
        ...stateObj.chunkQueue,
        [key]: [...existing, packet].slice(-25),
      };
      return { chunkQueue: nextQueue };
    }),

  triggerChunkAssembly: (hash = "unmapped") =>
    set((stateObj) => {
      const packets = stateObj.chunkQueue[hash];
      if (!packets || packets.length === 0) {
        return stateObj;
      }
      const nextQueue = { ...stateObj.chunkQueue };
      delete nextQueue[hash];
      return {
        chunkQueue: nextQueue,
        activeAssembly: {
          hash,
          packets,
          startedAt: Date.now(),
        },
      };
    }),

  clearActiveAssembly: () => set({ activeAssembly: null }),

  setLocalNodeIp: (ip) =>
    set((stateObj) => {
      const sanitized = ip.trim() || "64.31.48.109";
      const existing = stateObj.nodes.find((node) => node.isLocal);
      if (!existing) {
        return { localNodeIp: sanitized };
      }
      const geo = approximateGeoFromIp(sanitized);
      const updatedLocal = {
        ...existing,
        ip: sanitized,
        latitude: geo.latitude,
        longitude: geo.longitude,
        position: [geo.longitude, 0, 0] as [number, number, number],
        lastActivity: new Date().toISOString(),
      };
      return {
        localNodeIp: sanitized,
        nodes: stateObj.nodes.map((node) =>
          node.id === existing.id ? updatedLocal : node,
        ),
      };
    }),

  pushRouterEvent: (event) =>
    set((stateObj) => {
      const events = [event, ...stateObj.routerEvents]
        .sort((a, b) => b.timestamp - a.timestamp)
        .slice(0, 30);
      const selected =
        stateObj.selectedRouterEventId ?? event.id;
      return {
        routerEvents: events,
        selectedRouterEventId: selected,
      };
    }),

  setSelectedRouterEvent: (id) => set({ selectedRouterEventId: id }),

  setDetailDrawerOpen: (open) => set({ detailDrawerOpen: open }),
}));
