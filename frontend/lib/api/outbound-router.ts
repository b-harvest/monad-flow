import { z } from "zod";

const TimestampSchema = z.union([z.string(), z.number(), z.date()]);
const MessageTypeSchema = z
  .union([z.number(), z.string()])
  .transform((value) => {
    if (typeof value === "number") {
      return value;
    }
    const parsed = Number.parseInt(value, 10);
    return Number.isNaN(parsed) ? 0 : parsed;
  });

const CamelCaseNetworkVersionSchema = z.object({
  serializeVersion: z.number(),
  compressionVersion: z.number(),
});

const PascalCaseNetworkVersionSchema = z.object({
  SerializeVersion: z.number(),
  CompressionVersion: z.number(),
});

export type NetworkMessageVersion = z.infer<
  typeof CamelCaseNetworkVersionSchema
>;

export const NetworkMessageVersionSchema = z
  .union([CamelCaseNetworkVersionSchema, PascalCaseNetworkVersionSchema])
  .transform((value): NetworkMessageVersion => ({
    serializeVersion:
      "serializeVersion" in value
        ? value.serializeVersion
        : value.SerializeVersion,
    compressionVersion:
      "compressionVersion" in value
        ? value.compressionVersion
        : value.CompressionVersion,
  }));

export const OutboundRouterEventSchema = z
  .object({
    _id: z.string(),
    __v: z.number().optional(),
    type: z.string().optional(),
    version: NetworkMessageVersionSchema,
    messageType: MessageTypeSchema,
    data: z.unknown().optional(),
    appMessageHash: z.string().optional(),
    timestamp: TimestampSchema,
  })
  .passthrough();

export type OutboundRouterEvent = z.infer<typeof OutboundRouterEventSchema>;
export type RouterLogEntry = OutboundRouterEvent;

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
    const result = OutboundRouterEventSchema.safeParse(entry);
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

export async function fetchOutboundAppMessage(
  id: string,
): Promise<OutboundRouterEvent> {
  if (!id) {
    throw new Error("App message ID is required");
  }
  const response = await fetch(`${API_BASE}/api/app-message/${id}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch app message (${response.status})`);
  }
  const json = await response.json();
  const result = OutboundRouterEventSchema.safeParse(json);
  if (!result.success) {
    throw new Error(
      `Invalid app message payload: ${result.error.issues
        .map((issue) => issue.message)
        .join(", ")}`,
    );
  }
  return result.data;
}
