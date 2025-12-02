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
    batchIngestPings: (items: { ip: string; rtt_ms?: number }[]) => void;
    ipToPubkey: Record<string, string>;
    recentPings: {
      id: string;
      ip: string;
      rtt_ms: number;
      timestamp: number;
      name: string;
    }[];
  };

export const useNodePulseStore = create<NodePulseState>()((...a) => ({
  ...createNodeSlice(...a),
  ...createRouterSlice(...a),
  ...createTelemetrySlice(...a),
  ...createNetworkSlice(...a),
  ...createNetworkSlice(...a),
  ...createPlaybackSlice(...a),
  ipToPubkey: {},
  recentPings: [],

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

    // 0. Update IP->Pubkey map from unicast chunks
    const nextIpToPubkey = { ...state.ipToPubkey };
    let mapUpdated = false;

    items.forEach((item) => {
      const payload = item.packet.payload;
      // Check for unicast chunks (broadcast flags are falsey)
      // Note: broadcast is optional, so undefined means false (not broadcast)
      const isBroadcast =
        payload.broadcast || payload.broadCast || payload.secondaryBroadcast;

      if (
        payload &&
        !isBroadcast &&
        payload.secp_pubkey &&
        payload.network?.ipv4?.srcIp
      ) {
        const ip = payload.network.ipv4.srcIp;
        if (nextIpToPubkey[ip] !== payload.secp_pubkey) {
          nextIpToPubkey[ip] = payload.secp_pubkey;
          mapUpdated = true;
        }
      }
    });

    // 1. Process Peers (Deduplicate)
    const newPeers = new Map<string, { ip: string; port: number }>();
    items.forEach((item) => {
      const key = `chunk-${item.peer.ip}:${item.peer.port}`;
      if (!state.nodes.find((n) => n.id === key)) {
        newPeers.set(key, item.peer);
      }
    });

    let nextNodes = state.nodes;

    // If map updated, refresh names for existing nodes
    if (mapUpdated) {
      nextNodes = nextNodes.map((node) => {
        const pubKey = nextIpToPubkey[node.ip];
        // If we found a pubKey and the node doesn't have one or name needs update
        if (pubKey && (!node.pubKey || node.pubKey !== pubKey)) {
          // Extract port from ID (format: chunk-IP:PORT)
          const portStr = node.id.split(":")[1];
          const port = portStr ? parseInt(portStr) : 0;
          return {
            ...node,
            pubKey,
            name: resolvePeerName(pubKey, node.ip, port),
          };
        }
        return node;
      });
    }

    if (newPeers.size > 0) {
      const newNodes = Array.from(newPeers.values()).map((p) => {
        const geo = approximateGeoFromIp(p.ip, p.port);
        const pubKey = nextIpToPubkey[p.ip];
        return {
          id: `chunk-${p.ip}:${p.port}`,
          name: resolvePeerName(pubKey, p.ip, p.port),
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
          pubKey: pubKey,
        } as any; // Cast to MonadNode
      });
      nextNodes = [...nextNodes, ...newNodes];
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
      ipToPubkey: mapUpdated ? nextIpToPubkey : state.ipToPubkey,
    });
  },

  batchIngestPings: (items) => {
    const [set, get] = a;
    const state = get();
    const updates = new Map<string, number>();
    const now = Date.now();
    const newPings: {
      id: string;
      ip: string;
      rtt_ms: number;
      timestamp: number;
      name: string;
    }[] = [];

    items.forEach((item) => {
      if (item.rtt_ms !== undefined) {
        updates.set(item.ip, item.rtt_ms);

        // Resolve name
        const pubKey = state.ipToPubkey[item.ip];
        // We don't have port info in ping, assume default or just use IP for fallback
        // resolvePeerName expects port, let's pass 0 if unknown
        const name = resolvePeerName(pubKey, item.ip, 0);

        newPings.push({
          id: `ping-${item.ip}-${now}-${Math.random().toString(36).slice(2, 6)}`,
          ip: item.ip,
          rtt_ms: item.rtt_ms,
          timestamp: now,
          name,
        });
      }
    });

    if (updates.size === 0) return;

    const nextRecentPings = [...newPings, ...state.recentPings].slice(0, 50);

    set({
      nodes: state.nodes.map((node) => {
        const newLatency = updates.get(node.ip);
        if (newLatency !== undefined) {
          return { ...node, latency: newLatency };
        }
        return node;
      }),
      recentPings: nextRecentPings,
    });
  },
}));
