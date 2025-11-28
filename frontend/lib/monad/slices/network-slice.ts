import { StateCreator } from "zustand";
import type { ChunkPacketRecord } from "@/types/monad";

const ensureRecentHashes = (current: string[], hash?: string) => {
  if (!hash) {
    return current;
  }
  const next = [...current, hash];
  if (next.length > 7) {
    next.shift();
  }
  return next;
};

export interface NetworkSlice {
  chunkQueue: Record<string, ChunkPacketRecord[]>;
  recentAppMessageHashes: string[];
  activeAssembly: {
    hash: string;
    packets: ChunkPacketRecord[];
    startedAt: number;
  } | null;
  pushChunkPacket: (packet: ChunkPacketRecord) => void;
  triggerChunkAssembly: (hash?: string) => void;
  clearActiveAssembly: () => void;
  resetNetworkData: () => void;
}

export const createNetworkSlice: StateCreator<NetworkSlice> = (set) => ({
  chunkQueue: {},
  recentAppMessageHashes: [],
  activeAssembly: null,

  pushChunkPacket: (packet) =>
    set((state) => {
      const key = packet.appMessageHash ?? "unmapped";
      const existing = state.chunkQueue[key] ?? [];
      const nextQueue = {
        ...state.chunkQueue,
        [key]: [...existing, packet].slice(-50),
      };
      return {
        chunkQueue: nextQueue,
        recentAppMessageHashes: ensureRecentHashes(
          state.recentAppMessageHashes,
          packet.appMessageHash,
        ),
      };
    }),

  triggerChunkAssembly: (hash = "unmapped") =>
    set((state) => {
      const packets = state.chunkQueue[hash];
      if (!packets || packets.length === 0) {
        return state;
      }
      const nextQueue = { ...state.chunkQueue };
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

  resetNetworkData: () =>
    set({
      chunkQueue: {},
      activeAssembly: null,
    }),
});
