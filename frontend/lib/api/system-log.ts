import { z } from "zod";

const SystemLogSchema = z.object({
  _id: z.string(),
  unit: z.string(),
  message: z.string(),
  timestamp: z.string(),
  pid: z.string().optional(),
  __v: z.number().optional(),
});

export type SystemLogEvent = z.infer<typeof SystemLogSchema>;

export type SystemLogSafeParseResult =
  | { success: true; data: SystemLogEvent }
  | { success: false; error: Error };

export const SystemLogEventSchema = {
  parse: (payload: unknown): SystemLogEvent => SystemLogSchema.parse(payload),
  safeParse: (payload: unknown): SystemLogSafeParseResult => {
    try {
      return { success: true, data: SystemLogSchema.parse(payload) };
    } catch (error) {
      return { success: false, error: error as Error };
    }
  },
};
