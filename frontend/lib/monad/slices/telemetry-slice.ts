import { StateCreator } from "zustand";
import type { BpfTraceEvent } from "@/lib/api/bpf-trace";
import type { SystemLogEvent } from "@/lib/api/system-log";
import type { SchedulerEvent } from "@/lib/api/scheduler";
import type { PerfStatEvent } from "@/lib/api/perf-stat";
import type { OffCpuEvent } from "@/lib/api/off-cpu";
import type { TurboStatEvent } from "@/lib/api/turbo-stat";

export interface TelemetrySlice {
  bpfTraceEvents: Record<string, BpfTraceEvent[]>;
  systemLogEvents: Record<string, SystemLogEvent[]>;
  schedulerEvents: Record<string, SchedulerEvent>;
  perfStatEvents: Record<string, PerfStatEvent>;
  offCpuEvents: Record<string, OffCpuEvent>;
  turboStatEvents: TurboStatEvent[];
  pushBpfTraceEvent: (event: BpfTraceEvent) => void;
  pushSystemLogEvent: (event: SystemLogEvent) => void;
  pushSchedulerEvent: (event: SchedulerEvent) => void;
  pushPerfStatEvent: (event: PerfStatEvent) => void;
  pushOffCpuEvent: (event: OffCpuEvent) => void;
  pushTurboStatEvent: (event: TurboStatEvent) => void;
}

export const createTelemetrySlice: StateCreator<TelemetrySlice> = (set) => ({
  bpfTraceEvents: {},
  systemLogEvents: {},
  schedulerEvents: {},
  perfStatEvents: {},
  offCpuEvents: {},
  turboStatEvents: [],

  pushBpfTraceEvent: (event) =>
    set((state) => {
      const funcName = event.func_name;
      const current = state.bpfTraceEvents[funcName] ?? [];
      const next = [...current, event].slice(-20);
      return {
        bpfTraceEvents: {
          ...state.bpfTraceEvents,
          [funcName]: next,
        },
      };
    }),

  pushSystemLogEvent: (event) =>
    set((state) => {
      const unit = event.unit;
      const current = state.systemLogEvents[unit] ?? [];
      const next = [...current, event].slice(-10);
      return {
        systemLogEvents: {
          ...state.systemLogEvents,
          [unit]: next,
        },
      };
    }),

  pushSchedulerEvent: (event) =>
    set((state) => {
      const pid = event.pid ?? event.main_pid;
      return {
        schedulerEvents: {
          ...state.schedulerEvents,
          [pid]: event,
        },
      };
    }),

  pushPerfStatEvent: (event) =>
    set((state) => ({
      perfStatEvents: {
        ...state.perfStatEvents,
        [event.pid]: event,
      },
    })),

  pushOffCpuEvent: (event) =>
    set((state) => ({
      offCpuEvents: {
        ...state.offCpuEvents,
        [event.pid]: event,
      },
    })),

  pushTurboStatEvent: (event) =>
    set(() => ({
      turboStatEvents: [event],
    })),
});
