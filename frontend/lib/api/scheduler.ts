import { z } from "zod";

export const SchedulerEventSchema = z.object({
  _id: z.string(),
  timestamp: z.string(),
  main_pid: z.string(),
  tid: z.string(),
  thread_name: z.string(),
  wait_delta_ms: z.number(),
  run_delta_ms: z.number(),
  ctx_switches: z.number(),
  __v: z.number().optional(),
});

export type SchedulerEvent = z.infer<typeof SchedulerEventSchema>;

export type SchedulerSafeParseResult =
  | { success: true; data: SchedulerEvent }
  | { success: false; error: Error };

export const SchedulerParser = {
  parse: (payload: unknown): SchedulerEvent =>
    SchedulerEventSchema.parse(payload),
  safeParse: (payload: unknown): SchedulerSafeParseResult => {
    try {
      return { success: true, data: SchedulerEventSchema.parse(payload) };
    } catch (error) {
      return { success: false, error: error as Error };
    }
  },
};
