import { z } from "zod";

export const TurboStatEventSchema = z.object({
  _id: z.string().optional(),
  timestamp: z.string(),
  core: z.string(),
  cpu: z.string(),
  avg_mhz: z.number(),
  busy_pct: z.number(),
  bzy_mhz: z.number(),
  tsc_mhz: z.number(),
  ipc: z.number(),
  irq: z.number(),
  cor_watt: z.number(),
  pkg_watt: z.number(),
  __v: z.number().optional(),
});

export type TurboStatEvent = z.infer<typeof TurboStatEventSchema>;

export type TurboStatSafeParseResult =
  | { success: true; data: TurboStatEvent }
  | { success: false; error: Error };

export const TurboStatParser = {
  parse: (payload: unknown): TurboStatEvent =>
    TurboStatEventSchema.parse(payload),
  safeParse: (payload: unknown): TurboStatSafeParseResult => {
    try {
      return { success: true, data: TurboStatEventSchema.parse(payload) };
    } catch (error) {
      return { success: false, error: error as Error };
    }
  },
};
