import { StateCreator } from "zustand";
import type { OutboundRouterEvent } from "@/lib/api/outbound-router";
import {
  getProposalSnapshot,
  type ProposalSnapshot,
} from "@/lib/monad/normalize-proposal";

const MAX_ROUTER_EVENTS = 50;
const MAX_PROPOSAL_SNAPSHOTS = 20;
const MAX_FORWARDED_TX = 25;

const parseTimestampMs = (value: string | number | Date | undefined) => {
  if (typeof value === "number") {
    return value;
  }
  if (typeof value === "string") {
    const fromNumber = Number(value);
    if (!Number.isNaN(fromNumber)) {
      return fromNumber;
    }
    const fromDate = Date.parse(value);
    return Number.isNaN(fromDate) ? 0 : fromDate;
  }
  if (value instanceof Date) {
    return value.getTime();
  }
  return 0;
};

export interface ForwardedTxSummary {
  hash: string;
  to?: string;
  value?: string;
  timestamp: number;
}

export interface RouterSlice {
  routerEvents: OutboundRouterEvent[];
  selectedRouterEventId: string | null;
  proposalSnapshots: ProposalSnapshot[];
  forwardedTxs: ForwardedTxSummary[];
  pushRouterEvent: (event: OutboundRouterEvent) => void;
  clearRouterEvents: () => void;
  setSelectedRouterEvent: (id: string) => void;
}

const getTypeId = (value: Record<string, any> | undefined | null) => {
  if (!value) return null;
  const raw = value.typeId ?? value.TypeID ?? value.type ?? value.Type;
  if (raw === undefined) return null;
  const num = typeof raw === "number" ? raw : Number(raw);
  return Number.isFinite(num) ? num : null;
};

export const createRouterSlice: StateCreator<RouterSlice> = (set) => ({
  routerEvents: [],
  selectedRouterEventId: null,
  proposalSnapshots: [],
  forwardedTxs: [],

  pushRouterEvent: (event) =>
    set((state) => {
      // Optimized: Prepend and slice, assuming mostly ordered arrival. Limit to 50.
      const events = [event, ...state.routerEvents].sort(
        (a, b) =>
          parseTimestampMs(b.timestamp) - parseTimestampMs(a.timestamp),
      );
      const trimmed = events.slice(0, MAX_ROUTER_EVENTS);
      const selected = state.selectedRouterEventId ?? event._id;

      const newProposal = getProposalSnapshot(event);
      let nextSnapshots = state.proposalSnapshots;

      if (newProposal) {
        // 1. Append new proposal
        const combined = [...state.proposalSnapshots, newProposal];
        
        // 2. Deduplicate by round (keep the one with more info or just the latest processed? 
        //    Usually round is unique. We'll use a Map to keep unique rounds.)
        const uniqueMap = new Map<number, ProposalSnapshot>();
        combined.forEach(p => uniqueMap.set(p.round, p));
        
        // 3. Sort by round ascending
        const sorted = Array.from(uniqueMap.values()).sort((a, b) => a.round - b.round);
        
        // 4. Slice to keep only the last N
        nextSnapshots = sorted.slice(-MAX_PROPOSAL_SNAPSHOTS);
      }

      // Extract ForwardedTxs if present (AppMessage - ForwardedTxs)
      let nextForwarded = state.forwardedTxs;
      if (event.messageType === 1) {
        const data =
          event.data && typeof event.data === "object"
            ? (event.data as Record<string, unknown>)
            : null;
        const typeId = getTypeId(data ?? undefined);
        if (typeId === 4) {
          const payload = (data?.payload ?? []) as unknown;
          if (Array.isArray(payload)) {
            const ts = parseTimestampMs(event.timestamp);
            const extracted: ForwardedTxSummary[] = payload
              .map((raw) => {
                if (
                  !raw ||
                  typeof raw !== "object" ||
                  !("hash" in raw)
                ) {
                  return null;
                }
                const tx = raw as {
                  hash?: string;
                  to?: string;
                  value?: string;
                };
                if (!tx.hash) return null;
                return {
                  hash: tx.hash,
                  to: tx.to,
                  value: tx.value,
                  timestamp: ts,
                };
              })
              .filter(
                (item): item is ForwardedTxSummary =>
                  item !== null,
              );
            if (extracted.length > 0) {
              nextForwarded = [
                ...extracted,
                ...state.forwardedTxs,
              ].slice(0, MAX_FORWARDED_TX);
            }
          }
        }
      }

      return {
        routerEvents: trimmed,
        selectedRouterEventId: selected,
        proposalSnapshots: nextSnapshots,
        forwardedTxs: nextForwarded,
      };
    }),

  clearRouterEvents: () =>
    set(() => ({
      routerEvents: [],
      selectedRouterEventId: null,
    })),

  setSelectedRouterEvent: (id) => set({ selectedRouterEventId: id }),
});
