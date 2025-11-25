import { z } from "zod";

const Ipv4Schema = z.object({
  srcIp: z.string(),
  dstIp: z.string(),
  protocol: z.string(),
});

const PortSchema = z.object({
  srcPort: z.number(),
  dstPort: z.number(),
});

const NetworkSchema = z.object({
  ipv4: Ipv4Schema,
  port: PortSchema,
});

const MerkleProofEntry = z.union([z.string(), z.array(z.string())]);

export const MonadChunkEventSchema = z.object({
  _id: z.string().optional(),
  network: NetworkSchema,
  signature: z.string(),
  version: z.number(),
  flags: z.number(),
  broadcast: z.boolean().optional(),
  broadCast: z.boolean().optional(),
  secondaryBroadcast: z.boolean(),
  merkleTreeDepth: z.number(),
  epoch: z.string(),
  timestampMs: z.string(),
  appMessageHash: z.string(),
  appMessageLen: z.number(),
  merkleProof: z.array(MerkleProofEntry),
  firstHopRecipient: z.string(),
  merkleLeafIdx: z.number(),
  reserved: z.number(),
  chunkId: z.number(),
  timestamp: z.string(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
  __v: z.number().optional(),
});

export type MonadChunkEvent = z.infer<typeof MonadChunkEventSchema>;

export type MonadChunkSafeParseResult =
  | { success: true; data: MonadChunkEvent }
  | { success: false; error: Error };

export const MonadChunkParser = {
  parse: (payload: unknown): MonadChunkEvent =>
    MonadChunkEventSchema.parse(payload),
  safeParse: (payload: unknown): MonadChunkSafeParseResult => {
    try {
      return { success: true, data: MonadChunkEventSchema.parse(payload) };
    } catch (error) {
      return { success: false, error: error as Error };
    }
  },
};
