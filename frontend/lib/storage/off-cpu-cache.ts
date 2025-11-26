"use client";

import { z } from "zod";
import type { OffCpuEvent } from "@/lib/api/off-cpu";
import { OffCpuParser } from "@/lib/api/off-cpu";

const STORAGE_KEY = "monad-flow:off-cpu";
export const MAX_OFF_CPU_ENTRIES = 30;

const EventArraySchema = z.array(
  z.object({
    _id: z.string(),
    timestamp: z.string(),
    process_name: z.string(),
    pid: z.string(),
    tid: z.string(),
    duration_us: z.number(),
    stack: z.array(z.string()),
    __v: z.number().optional(),
  }),
);
type Listener = () => void;
const listeners = new Set<Listener>();

let hydratedBuffer: OffCpuEvent[] = [];
let isHydrated = false;

function loadFromStorage(): OffCpuEvent[] {
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

function persistBuffer(next: OffCpuEvent[]) {
  hydratedBuffer = next;
  if (typeof window !== "undefined") {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch (error) {
      console.warn("[OFF_CPU] Failed to persist cache:", error);
    }
  }
  listeners.forEach((listener) => {
    try {
      listener();
    } catch (error) {
      console.warn("[OFF_CPU] Listener error:", error);
    }
  });
}

export function getOffCpuEvents() {
  return loadFromStorage();
}

export function appendOffCpuEvent(payload: unknown) {
  const parsed = OffCpuParser.safeParse(payload);
  if (!parsed.success) {
    console.error("[OFF_CPU] Parse error:", parsed.error, payload);
    throw parsed.error;
  }
  const current = loadFromStorage();
  const next = [...current, parsed.data].slice(-MAX_OFF_CPU_ENTRIES);
  persistBuffer(next);
  return { event: parsed.data, snapshot: next };
}

export function clearOffCpuEvents() {
  persistBuffer([]);
}

export function subscribeToOffCpuEvents(listener: Listener) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}
