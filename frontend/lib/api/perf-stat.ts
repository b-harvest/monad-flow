import { z } from "zod";

export const PerfMetricSchema = z.object({
  event: z.string(),
  value: z.string(),
  unit: z.string().optional(),
  metric_val: z.string().optional(),
  run_pct: z.string().optional(),
});

export const PerfStatEventSchema = z.object({
  _id: z.string(),
  timestamp: z.string(),
  perf_timestamp: z.string(),
  pid: z.string(),
  metrics: z.array(PerfMetricSchema),
  __v: z.number().optional(),
});

export type PerfStatEvent = z.infer<typeof PerfStatEventSchema>;

export type PerfStatSafeParseResult =
  | { success: true; data: PerfStatEvent }
  | { success: false; error: Error };

export const PerfStatParser = {
  parse: (payload: unknown): PerfStatEvent => PerfStatEventSchema.parse(payload),
  safeParse: (payload: unknown): PerfStatSafeParseResult => {
    try {
      return { success: true, data: PerfStatEventSchema.parse(payload) };
    } catch (error) {
      return { success: false, error: error as Error };
    }
  },
};
