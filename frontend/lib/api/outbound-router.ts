import { z } from "zod";

const NetworkMessageVersionSchema = z.object({
  serializeVersion: z.number(),
  compressionVersion: z.number(),
});

const PeerDiscoveryMessageSchema = z.object({
  version: z.number(),
  type: z.number(),
  payload: z.record(z.any()).optional(),
});

const FullNodesGroupMessageSchema = z.object({
  version: z.number(),
  type: z.number(),
  payload: z.record(z.any()).optional(),
});

const MonadVersionSchema = z.object({
  protocolVersion: z.number(),
  clientVersionMajor: z.number(),
  clientVersionMinor: z.number(),
  hashVersion: z.number(),
  serializeVersion: z.number(),
});

const MonadMessageSchema = z.object({
  version: MonadVersionSchema,
  typeId: z.number(),
  payload: z.record(z.any()),
});

const OutboundRouterCombinedSchema = z.object({
  version: NetworkMessageVersionSchema,
  messageType: z.number(),
  peerDiscovery: PeerDiscoveryMessageSchema.optional(),
  fullNodesGroup: FullNodesGroupMessageSchema.optional(),
  appMessage: MonadMessageSchema.optional(),
});

const RouterLogEntrySchema = z.object({
  _id: z.string(),
  type: z.string(),
  appMessageHash: z.string().optional(),
  data: OutboundRouterCombinedSchema.optional(),
  timestamp: z.union([z.string(), z.number()]),
});

export type RouterLogEntry = z.infer<typeof RouterLogEntrySchema>;

export interface FetchRouterLogsOptions {
  from?: string;
  to?: string;
  limit?: number;
  windowMs?: number;
}

const API_BASE = process.env.BACKEND_URL ?? "http://51.195.24.236:3000";

export async function fetchRouterLogs(
  options: FetchRouterLogsOptions = {},
) {
  const now = Date.now();
  const to = options.to ?? new Date(now).toISOString();
  const from =
    options.from ??
    new Date(now - (options.windowMs ?? 1000 * 60 * 2)).toISOString();
  const limitParam = options.limit ? `&limit=${options.limit}` : "";
  const response = await fetch(
    `${API_BASE}/api/logs/router?from=${encodeURIComponent(from)}&to=${encodeURIComponent(
      to,
    )}${limitParam}`,
  );
  if (!response.ok) {
    throw new Error(`Failed to fetch router logs (${response.status})`);
  }
  const json = await response.json();
  if (!Array.isArray(json)) {
    throw new Error("Unexpected router log payload");
  }

  const parsed: RouterLogEntry[] = [];
  const issues: string[] = [];
  json.forEach((entry, index) => {
    const result = RouterLogEntrySchema.safeParse(entry);
    if (result.success) {
      parsed.push(result.data);
    } else {
      issues.push(
        `Entry ${index} failed validation: ${result.error.issues
          .map((issue) => issue.message)
          .join(", ")}`,
      );
    }
  });

  return { entries: parsed, issues };
}
