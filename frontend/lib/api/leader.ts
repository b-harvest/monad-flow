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

const API_BASE =
  process.env.BACKEND_URL ?? "http://51.195.24.236:3000";

export async function fetchLeaderSchedule(
  round: number,
  range = 5,
): Promise<LeaderEvent[]> {
  const params = new URLSearchParams({
    range: String(range),
  });
  const response = await fetch(
    `${API_BASE}/api/leader/${round}?${params.toString()}`,
    {
      cache: "no-store",
    },
  );
  if (!response.ok) {
    throw new Error(
      `Failed to fetch leader schedule (${response.status})`,
    );
  }
  const json = await response.json();
  if (!Array.isArray(json)) {
    throw new Error("Unexpected leader schedule payload");
  }

  const result: LeaderEvent[] = [];
  json.forEach((entry) => {
    const parsed = LeaderEventSchema.safeParse(entry);
    if (parsed.success) {
      result.push(parsed.data);
    } else {
      console.warn(
        "[LEADER_API] Invalid leader entry",
        parsed.error,
      );
    }
  });

  return result;
}
