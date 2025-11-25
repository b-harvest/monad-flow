"use client";

import { z } from "zod";
import type { PerfStatEvent } from "@/lib/api/perf-stat";
import { PerfStatParser } from "@/lib/api/perf-stat";

const STORAGE_KEY = "monad-flow:perf-stat";
export const MAX_PERF_STAT_ENTRIES = 100;

const EventArraySchema = z.array(
  z.object({
    _id: z.string(),
    timestamp: z.string(),
    perf_timestamp: z.string(),
    pid: z.string(),
    metrics: z.array(
      z.object({
        event: z.string(),
        value: z.string(),
        unit: z.string().optional(),
        metric_val: z.string().optional(),
        run_pct: z.string().optional(),
      }),
    ),
    __v: z.number().optional(),
  }),
);

let hydratedBuffer: PerfStatEvent[] = [];
let isHydrated = false;

function loadFromStorage(): PerfStatEvent[] {
  if (typeof window === "undefined") {
    return hydratedBuffer;
  }
  if (isHydrated) {
    return hydratedBuffer;
  }
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    hydratedBuffer = raw ? EventArraySchema.parse(JSON.parse(raw)) : [];
  } catch {
    hydratedBuffer = [];
  } finally {
    isHydrated = true;
  }
  return hydratedBuffer;
}

function persistBuffer(next: PerfStatEvent[]) {
  hydratedBuffer = next;
  if (typeof window !== "undefined") {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch (error) {
      console.warn("[PERF_STAT] Failed to persist cache:", error);
    }
  }
}

export function getPerfStatEvents() {
  return loadFromStorage();
}

export function appendPerfStatEvent(payload: unknown) {
  const parsed = PerfStatParser.safeParse(payload);
  if (!parsed.success) {
    console.error("[PERF_STAT] Parse error:", parsed.error, payload);
    throw parsed.error;
  }
  const current = loadFromStorage();
  const next = [...current, parsed.data].slice(-MAX_PERF_STAT_ENTRIES);
  persistBuffer(next);
  return { event: parsed.data, snapshot: next };
}

export function clearPerfStatEvents() {
  persistBuffer([]);
}
