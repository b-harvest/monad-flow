import { StateCreator } from "zustand";
import type {
  AlertToast,
  HistoricalEvent,
  MonitoringEvent,
  PlaybackState,
  PulseVisualEffect,
} from "@/types/monad";

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

export interface PlaybackSlice {
  eventLog: MonitoringEvent[];
  visualEffects: PulseVisualEffect[];
  playback: PlaybackState;
  alert: AlertToast | null;
  lastEventTimestamp: number;
  historicalTimeline: HistoricalEvent[];
  historicalPointer: number;
  detailDrawerOpen: boolean;
  pushEvent: (event: MonitoringEvent) => void;
  addEffect: (effect: PulseVisualEffect) => void;
  pruneEffects: (now?: number) => void;
  setPlayback: (patch: Partial<PlaybackState>) => void;
  setAlert: (toast: AlertToast | null) => void;
  setDetailDrawerOpen: (open: boolean) => void;
  setHistoricalTimeline: (timeline: HistoricalEvent[]) => void;
  setHistoricalPointer: (index: number) => void;
  resetHistoricalTimeline: () => void;
}

export const createPlaybackSlice: StateCreator<PlaybackSlice> = (set) => ({
  eventLog: [],
  visualEffects: [],
  playback: initialPlayback,
  alert: null,
  lastEventTimestamp: 0,
  historicalTimeline: [],
  historicalPointer: 0,
  detailDrawerOpen: false,

  pushEvent: (event) =>
    set((state) => {
      const next = [...state.eventLog, event];
      next.sort((a, b) => b.timestamp - a.timestamp);
      const trimmed = next.slice(0, 50);
      return {
        eventLog: trimmed,
        lastEventTimestamp: event.timestamp,
      };
    }),

  addEffect: (effect) =>
    set((state) => ({
      visualEffects: [...state.visualEffects, effect],
    })),

  pruneEffects: (now = Date.now()) =>
    set((state) => ({
      visualEffects: state.visualEffects.filter(
        (effect) => effect.createdAt + effect.ttl > now,
      ),
    })),

  setPlayback: (patch) =>
    set((state) => ({
      playback: { ...state.playback, ...patch },
    })),

  setAlert: (toast) => set({ alert: toast }),

  setDetailDrawerOpen: (open) => set({ detailDrawerOpen: open }),

  setHistoricalTimeline: (timeline) =>
    set({
      historicalTimeline: timeline,
      historicalPointer: 0,
    }),

  setHistoricalPointer: (index) => set({ historicalPointer: index }),

  resetHistoricalTimeline: () =>
    set({
      historicalTimeline: [],
      historicalPointer: 0,
    }),
});
