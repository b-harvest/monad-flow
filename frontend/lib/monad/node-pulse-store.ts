import { create } from "zustand";
import { createNodeSlice, type NodeSlice } from "./slices/node-slice";
import { createRouterSlice, type RouterSlice } from "./slices/router-slice";
import {
  createTelemetrySlice,
  type TelemetrySlice,
} from "./slices/telemetry-slice";
import { createNetworkSlice, type NetworkSlice } from "./slices/network-slice";
import {
  createPlaybackSlice,
  type PlaybackSlice,
} from "./slices/playback-slice";
import { approximateGeoFromIp } from "@/lib/geo/ip-to-geo";
import { resolvePeerName } from "@/lib/monad/known-peers";

export type NodePulseState = NodeSlice &
  RouterSlice &
  TelemetrySlice &
  NetworkSlice &
  PlaybackSlice & {
    batchIngestChunks: (items: any[]) => void;
  };

export const useNodePulseStore = create<NodePulseState>()((...a) => ({
  ...createNodeSlice(...a),
  ...createRouterSlice(...a),
  ...createTelemetrySlice(...a),
  ...createNetworkSlice(...a),
  ...createPlaybackSlice(...a),

  // Override resetNetworkGraph to clear both Node and Network slices
  // Override resetNetworkGraph to clear both Node and Network slices
  resetNetworkGraph: () => {
    const [set, get] = a;
    // Call reset logic from NodeSlice
    set({
      nodes: [],
      selectedNodeId: null,
    });
    // Call reset logic from NetworkSlice
    get().resetNetworkData();
  },

  batchIngestChunks: (items: any[]) => {
    const [set, get] = a;
    const state = get();
    const now = Date.now();

    // 1. Process Peers (Deduplicate)
    const newPeers = new Map<string, { ip: string; port: number }>();
    items.forEach((item) => {
      const key = `chunk-${item.peer.ip}:${item.peer.port}`;
      if (!state.nodes.find((n) => n.id === key)) {
        newPeers.set(key, item.peer);
      }
    });

    let nextNodes = state.nodes;
    if (newPeers.size > 0) {
      const newNodes = Array.from(newPeers.values()).map((p) => {
        // We need to import createChunkNode logic or duplicate it here.
        // Since createChunkNode is internal to node-slice, we can't easily access it.
        // For now, we'll use the upsertChunkPeer logic but batched.
        // Actually, let's just call upsertChunkPeer for new ones? No, that triggers re-renders.
        // We should duplicate the simple node creation logic here for performance.
        const geo = approximateGeoFromIp(p.ip, p.port);
        // To avoid importing heavy logic, let's just use a simplified node creation or rely on the fact that
        // approximateGeoFromIp is fast. We can import it if needed, but let's try to keep it simple.
        // Wait, `createChunkNode` in node-slice uses a cache.
        // Ideally, we should expose a `batchUpsertPeers` in NodeSlice.
        // But for this refactor, let's just create the nodes manually here.
        return {
          id: `chunk-${p.ip}:${p.port}`,
          name: resolvePeerName(p.ip, p.port),
          role: "validator",
          ip: p.ip,
          uptimePct: 92 + Math.random() * 6,
          participationRate: 85 + Math.random() * 10,
          lastActivity: new Date().toISOString(),
          state: "active",
          position: [geo.longitude, -0.5, 0], // CHUNK_RING_HEIGHT
          latitude: geo.latitude,
          longitude: geo.longitude,
          isLocal: false,
        } as any; // Cast to MonadNode
      });
      nextNodes = [...state.nodes, ...newNodes];
    }

    // 2. Process Effects
    const newEffects = items.map((item) => item.effect);
    const nextEffects = [...state.visualEffects, ...newEffects].slice(-200);

    // 3. Process Packets
    const nextChunkQueue = { ...state.chunkQueue };
    const nextRecentHashes = [...state.recentAppMessageHashes];

    items.forEach((item) => {
      const packet = item.packet;
      const key = packet.appMessageHash ?? "unmapped";
      const existing = nextChunkQueue[key] ?? [];
      nextChunkQueue[key] = [...existing, packet].slice(-50);

      if (packet.appMessageHash) {
        nextRecentHashes.push(packet.appMessageHash);
      }
    });

    // Limit recent hashes
    const limitedRecentHashes = nextRecentHashes.slice(-7);

    set({
      nodes: nextNodes,
      visualEffects: nextEffects,
      chunkQueue: nextChunkQueue,
      recentAppMessageHashes: limitedRecentHashes,
      lastEventTimestamp: now,
    });
  },
}));
