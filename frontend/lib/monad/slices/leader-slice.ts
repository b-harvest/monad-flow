import { StateCreator } from "zustand";
import type { LeaderEvent } from "@/lib/api/leader";

export interface LeaderSlice {
  leaders: LeaderEvent[];
  addLeader: (leader: LeaderEvent) => void;
}

export const createLeaderSlice: StateCreator<LeaderSlice> = (set) => ({
  leaders: [],
  addLeader: (leader) =>
    set((state) => {
      // Keep last 5 leaders.
      // Newest on the right (end of array).
      // If full, remove from left (start of array).
      // Logic: [...state.leaders, leader].slice(-5)
      // Wait, user said: "New leader added to right".
      // "When 5 full... add to left".
      // If I interpret "add to left" as "shift left", then slice(-5) is correct (keeps last 5).
      // If I interpret as "newest goes to left", then [leader, ...state.leaders].slice(0, 5).
      // Given "New leader added to right" initially, I will stick to Append + Slice.
      // Visual representation will handle the "left/right" placement if needed, but array order usually implies time.
      // Array: [Oldest, ..., Newest]
      
      const nextLeaders = [...state.leaders, leader];
      if (nextLeaders.length > 5) {
        return { leaders: nextLeaders.slice(-5) };
      }
      return { leaders: nextLeaders };
    }),
});
