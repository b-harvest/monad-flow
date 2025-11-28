import { StateCreator } from "zustand";
import type { OutboundRouterEvent } from "@/lib/api/outbound-router";
import {
  getProposalSnapshot,
  type ProposalSnapshot,
} from "@/lib/monad/normalize-proposal";

const MAX_ROUTER_EVENTS = 50;
const MAX_PROPOSAL_SNAPSHOTS = 20;

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

export interface RouterSlice {
  routerEvents: OutboundRouterEvent[];
  selectedRouterEventId: string | null;
  proposalSnapshots: ProposalSnapshot[];
  pushRouterEvent: (event: OutboundRouterEvent) => void;
  clearRouterEvents: () => void;
  setSelectedRouterEvent: (id: string) => void;
}

export const createRouterSlice: StateCreator<RouterSlice> = (set) => ({
  routerEvents: [],
  selectedRouterEventId: null,
  proposalSnapshots: [],

  pushRouterEvent: (event) =>
    set((state) => {
      // Optimized: Prepend and slice, assuming mostly ordered arrival. Limit to 50.
      const events = [event, ...state.routerEvents].sort(
        (a, b) =>
          parseTimestampMs(b.timestamp) - parseTimestampMs(a.timestamp),
      );
      const trimmed = events.slice(0, MAX_ROUTER_EVENTS);
      const selected = state.selectedRouterEventId ?? event._id;
      const proposal = getProposalSnapshot(event);
      const proposalSnapshots = proposal
        ? [...state.proposalSnapshots, proposal].slice(-MAX_PROPOSAL_SNAPSHOTS)
        : state.proposalSnapshots;
      return {
        routerEvents: trimmed,
        selectedRouterEventId: selected,
        proposalSnapshots,
      };
    }),

  clearRouterEvents: () =>
    set(() => ({
      routerEvents: [],
      selectedRouterEventId: null,
    })),

  setSelectedRouterEvent: (id) => set({ selectedRouterEventId: id }),
});
