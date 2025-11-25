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
} from "@/types/monad";
import {
  arrangeNodeOrbit,
  createValidatorNode,
  generateInitialNodes,
} from "./generate-node-layout";

interface NodePulseState {
  nodes: MonadNode[];
  metrics: ConsensusMetrics;
  eventLog: MonitoringEvent[];
  visualEffects: PulseVisualEffect[];
  playback: PlaybackState;
  alert: AlertToast | null;
  lastEventTimestamp: number;
  preferences: {
    autoRotate: boolean;
    showParticles: boolean;
  };
  selectedNodeId: string | null;
  nodeTelemetry: Record<string, NodeTelemetryDigest>;
  routerEvents: OutboundRouterEventSummary[];
  selectedRouterEventId: string | null;
  detailDrawerOpen: boolean;
  setNodeState: (nodeId: string, state: MonadNodeState) => void;
  rotateLeader: (nodeId: string) => void;
  pushEvent: (event: MonitoringEvent) => void;
  addEffect: (effect: PulseVisualEffect) => void;
  pruneEffects: (now?: number) => void;
  setMetrics: (metrics: Partial<ConsensusMetrics>) => void;
  patchNode: (nodeId: string, patch: Partial<MonadNode>) => void;
  setPlayback: (patch: Partial<PlaybackState>) => void;
  setAlert: (toast: AlertToast | null) => void;
  setPreferences: (
    patch: Partial<NodePulseState["preferences"]>,
  ) => void;
  setSelectedNode: (nodeId: string | null) => void;
  upsertTelemetry: (nodeId: string, payload: NodeTelemetryDigest) => void;
  addValidator: () => void;
  pushRouterEvent: (event: OutboundRouterEventSummary) => void;
  setSelectedRouterEvent: (id: string) => void;
  setDetailDrawerOpen: (open: boolean) => void;
}

const initialNodes = generateInitialNodes(12);

const initialMetrics: ConsensusMetrics = {
  epoch: 128,
  round: 512,
  leaderId: initialNodes[0]?.id ?? "validator-1",
  tps: 1324,
  blockHeight: 983_455,
  avgBlockTime: 1.1,
  networkHealth: 94,
  connectionStatus: "connected",
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

export const useNodePulseStore = create<NodePulseState>((set) => ({
  nodes: initialNodes,
  metrics: initialMetrics,
  eventLog: [],
  visualEffects: [],
  playback: initialPlayback,
  alert: null,
  lastEventTimestamp: Date.now(),
  preferences: {
    autoRotate: true,
    showParticles: true,
  },
  selectedNodeId: null,
  nodeTelemetry: {},
  routerEvents: [],
  selectedRouterEventId: null,
  detailDrawerOpen: false,

  setNodeState: (nodeId, state) =>
    set((stateObj) => ({
      nodes: stateObj.nodes.map((node) =>
        node.id === nodeId ? { ...node, state } : node,
      ),
    })),

  rotateLeader: (nodeId) =>
    set((stateObj) => ({
      nodes: arrangeNodeOrbit(
        stateObj.nodes.map((node) => {
          if (node.id === nodeId) {
            return {
              ...node,
              role: "leader",
              state: "leader",
              cluster: undefined,
              parentId: undefined,
            };
          }
          if (node.id === stateObj.metrics.leaderId) {
            return {
              ...node,
              role: "validator",
              state: node.state === "leader" ? "active" : node.state,
              cluster: node.cluster ?? "primary",
              parentId: nodeId,
            };
          }
          return node;
        }),
      ),
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

  setPreferences: (patch) =>
    set((stateObj) => ({
      preferences: { ...stateObj.preferences, ...patch },
    })),

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

  addValidator: () =>
    set((stateObj) => {
      const ids = stateObj.nodes.map((node) => {
        const [, idx] = node.id.split("-");
        return Number(idx) || 1;
      });
      const nextIndex = Math.max(...ids) + 1;
      const parentId =
        stateObj.nodes.find((node) => node.isLocal)?.id ??
        stateObj.nodes[0]?.id ??
        "validator-1";
      const newNode = createValidatorNode(nextIndex, parentId);
      return { nodes: arrangeNodeOrbit([...stateObj.nodes, newNode]) };
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
