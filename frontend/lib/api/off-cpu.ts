import { z } from "zod";

export const OffCpuEventSchema = z.object({
  _id: z.string(),
  timestamp: z.string(),
  process_name: z.string(),
  pid: z.string(),
  tid: z.string(),
  duration_us: z.number(),
  stack: z.array(z.string()),
  __v: z.number().optional(),
});

export type OffCpuEvent = z.infer<typeof OffCpuEventSchema>;

export type OffCpuSafeParseResult =
  | { success: true; data: OffCpuEvent }
  | { success: false; error: Error };

export const OffCpuParser = {
  parse: (payload: unknown): OffCpuEvent => OffCpuEventSchema.parse(payload),
  safeParse: (payload: unknown): OffCpuSafeParseResult => {
    try {
      return { success: true, data: OffCpuEventSchema.parse(payload) };
    } catch (error) {
      return { success: false, error: error as Error };
    }
  },
};
