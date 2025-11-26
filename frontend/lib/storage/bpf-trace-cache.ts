"use client";

import { z } from "zod";
import {
  BpfTraceEventSchema,
  type BpfTraceEvent,
} from "@/lib/api/bpf-trace";

const STORAGE_KEY = "monad-flow:bpf-trace";
export const MAX_BPF_TRACE_ENTRIES = 500;

const EventArraySchema = z.array(BpfTraceEventSchema);
type Listener = () => void;
const listeners = new Set<Listener>();

let hydratedBuffer: BpfTraceEvent[] = [];
let isHydrated = false;

function loadFromStorage(): BpfTraceEvent[] {
  if (typeof window === "undefined") {
    return hydratedBuffer;
  }
  if (isHydrated) {
    return hydratedBuffer;
  }
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      hydratedBuffer = [];
    } else {
      const parsed = JSON.parse(raw);
      hydratedBuffer = EventArraySchema.parse(parsed);
    }
  } catch (error) {
    console.warn("[MonadFlow] Failed to load BPF trace cache:", error);
    hydratedBuffer = [];
  } finally {
    isHydrated = true;
  }
  return hydratedBuffer;
}

function persistBuffer(next: BpfTraceEvent[]) {
  hydratedBuffer = next;
  if (typeof window !== "undefined") {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch (error) {
      console.warn("[MonadFlow] Failed to persist BPF trace cache:", error);
    }
  }
  listeners.forEach((listener) => {
    try {
      listener();
    } catch (error) {
      console.warn("[MonadFlow] BPF listener error", error);
    }
  });
}

export function getBpfTraceEvents(): BpfTraceEvent[] {
  return loadFromStorage();
}

export function appendBpfTraceEvent(payload: unknown) {
  const parsed = BpfTraceEventSchema.safeParse(payload);
  if (!parsed.success) {
    const issues = parsed.error.errors.map((issue) => issue.message).join(", ");
    const message = `[BPF_TRACE] Failed to parse payload: ${issues}`;
    console.error(message, payload);
    throw new Error(message);
  }
  const current = loadFromStorage();
  const next = [...current, parsed.data].slice(-MAX_BPF_TRACE_ENTRIES);
  persistBuffer(next);
  return { event: parsed.data, snapshot: next };
}

export function clearBpfTraceEvents() {
  persistBuffer([]);
}

export function subscribeToBpfTraceEvents(listener: Listener) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}
