import { StateCreator } from "zustand";
import type { LeaderEvent } from "@/lib/api/leader";

export interface LeaderSlice {
  leaders: LeaderEvent[];
  addLeader: (leader: LeaderEvent) => void;
  leaderSchedule: LeaderEvent[];
  leaderScheduleRound: number | null;
  setLeaderSchedule: (round: number, leaders: LeaderEvent[]) => void;
}

export const createLeaderSlice: StateCreator<LeaderSlice> = (set) => ({
  leaders: [],
  addLeader: (leader) =>
    set((state) => {
      const nextLeaders = [...state.leaders, leader];
      if (nextLeaders.length > 5) {
        return { leaders: nextLeaders.slice(-5) };
      }
      return { leaders: nextLeaders };
    }),
  leaderSchedule: [],
  leaderScheduleRound: null,
  setLeaderSchedule: (round, leaders) =>
    set(() => ({
      leaderScheduleRound: round,
      leaderSchedule: leaders,
    })),
});
