"use client";

import { z } from "zod";
import type { SchedulerEvent } from "@/lib/api/scheduler";
import { SchedulerParser } from "@/lib/api/scheduler";

const STORAGE_KEY = "monad-flow:scheduler";
export const MAX_SCHEDULER_ENTRIES = 100;

const EventArraySchema = z.array(
  z.object({
    _id: z.string(),
    timestamp: z.string(),
    main_pid: z.string(),
    tid: z.string(),
    thread_name: z.string(),
    wait_delta_ms: z.number(),
    run_delta_ms: z.number(),
    ctx_switches: z.string(),
    __v: z.number().optional(),
  }),
);

let hydratedBuffer: SchedulerEvent[] = [];
let isHydrated = false;

function loadFromStorage(): SchedulerEvent[] {
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

function persistBuffer(next: SchedulerEvent[]) {
  hydratedBuffer = next;
  if (typeof window !== "undefined") {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch (error) {
      console.warn("[SCHEDULER] Failed to persist cache:", error);
    }
  }
}

export function getSchedulerEvents() {
  return loadFromStorage();
}

export function appendSchedulerEvent(payload: unknown) {
  const parsed = SchedulerParser.safeParse(payload);
  if (!parsed.success) {
    console.error("[SCHEDULER] Parse error:", parsed.error, payload);
    throw parsed.error;
  }
  const current = loadFromStorage();
  const next = [...current, parsed.data].slice(-MAX_SCHEDULER_ENTRIES);
  persistBuffer(next);
  return { event: parsed.data, snapshot: next };
}

export function clearSchedulerEvents() {
  persistBuffer([]);
}
