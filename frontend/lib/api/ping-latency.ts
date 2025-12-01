import { z } from "zod";

export const PingLatencySchema = z.object({
  ip: z.string(),
  rtt_ms: z.number().optional(),
  timestamp: z.string().or(z.number()).optional(), // NestJS default might be ISO string or timestamp
});

export type PingLatencyEvent = z.infer<typeof PingLatencySchema>;
