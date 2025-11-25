"use client";

import { z } from "zod";
import type { SystemLogEvent } from "@/lib/api/system-log";
import { SystemLogEventSchema } from "@/lib/api/system-log";

const STORAGE_KEY = "monad-flow:system-log";
export const MAX_SYSTEM_LOG_ENTRIES = 100;

const EventArraySchema = z.array(
  z.object({
    _id: z.string(),
    unit: z.string(),
    message: z.string(),
    timestamp: z.string(),
    pid: z.string().optional(),
    __v: z.number().optional(),
  }),
);

let hydratedBuffer: SystemLogEvent[] = [];
let isHydrated = false;

function loadFromStorage(): SystemLogEvent[] {
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

function persistBuffer(next: SystemLogEvent[]) {
  hydratedBuffer = next;
  if (typeof window !== "undefined") {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch (error) {
      console.warn("[SYSTEM_LOG] Failed to persist cache:", error);
    }
  }
}

export function getSystemLogEvents() {
  return loadFromStorage();
}

export function appendSystemLogEvent(payload: unknown) {
  const parsed = SystemLogEventSchema.safeParse(payload);
  if (!parsed.success) {
    console.error("[SYSTEM_LOG] Parse error:", parsed.error, payload);
    throw parsed.error;
  }
  const current = loadFromStorage();
  const next = [...current, parsed.data].slice(-MAX_SYSTEM_LOG_ENTRIES);
  persistBuffer(next);
  return { event: parsed.data, snapshot: next };
}

export function clearSystemLogEvents() {
  persistBuffer([]);
}
