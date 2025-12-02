import { z } from "zod";

export const LeaderEventSchema = z.object({
  _id: z.string(),
  epoch: z.number(),
  round: z.number(),
  node_id: z.string(),
  cert_pubkey: z.string(),
  stake: z.string(),
  timestamp: z.string(),
});

export type LeaderEvent = z.infer<typeof LeaderEventSchema>;
