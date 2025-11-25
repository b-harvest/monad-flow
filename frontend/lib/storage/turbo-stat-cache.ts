"use client";

import { z } from "zod";
import type { TurboStatEvent } from "@/lib/api/turbo-stat";
import { TurboStatParser, TurboStatEventSchema } from "@/lib/api/turbo-stat";

const STORAGE_KEY = "monad-flow:turbo-stat";
export const MAX_TURBO_STAT_ENTRIES = 120;

const EventArraySchema = z.array(TurboStatEventSchema);

let hydratedBuffer: TurboStatEvent[] = [];
let isHydrated = false;

function loadFromStorage(): TurboStatEvent[] {
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

function persistBuffer(next: TurboStatEvent[]) {
  hydratedBuffer = next;
  if (typeof window !== "undefined") {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch (error) {
      console.warn("[TURBO_STAT] Failed to persist cache:", error);
    }
  }
}

export function getTurboStatEvents() {
  return loadFromStorage();
}

export function appendTurboStatEvent(payload: unknown) {
  const parsed = TurboStatParser.safeParse(payload);
  if (!parsed.success) {
    console.error("[TURBO_STAT] Parse error:", parsed.error, payload);
    throw parsed.error;
  }
  const current = loadFromStorage();
  const next = [...current, parsed.data].slice(-MAX_TURBO_STAT_ENTRIES);
  persistBuffer(next);
  return { event: parsed.data, snapshot: next };
}

export function clearTurboStatEvents() {
  persistBuffer([]);
}
