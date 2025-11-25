"use client";

import { z } from "zod";
import type { MonadChunkEvent } from "@/lib/api/monad-chunk";
import {
  MonadChunkParser,
  MonadChunkEventSchema,
} from "@/lib/api/monad-chunk";

const STORAGE_KEY = "monad-flow:monad-chunk";
export const MAX_MONAD_CHUNK_ENTRIES = 80;

const EventArraySchema = z.array(MonadChunkEventSchema);

let hydratedBuffer: MonadChunkEvent[] = [];
let isHydrated = false;

function loadFromStorage(): MonadChunkEvent[] {
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

function persistBuffer(next: MonadChunkEvent[]) {
  hydratedBuffer = next;
  if (typeof window !== "undefined") {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch (error) {
      console.warn("[MONAD_CHUNK] Failed to persist cache:", error);
    }
  }
}

export function getMonadChunkEvents() {
  return loadFromStorage();
}

export function appendMonadChunkEvent(payload: unknown) {
  const parsed = MonadChunkParser.safeParse(payload);
  if (!parsed.success) {
    console.error("[MONAD_CHUNK] Parse error:", parsed.error, payload);
    throw parsed.error;
  }
  const current = loadFromStorage();
  const next = [...current, parsed.data].slice(-MAX_MONAD_CHUNK_ENTRIES);
  persistBuffer(next);
  return { event: parsed.data, snapshot: next };
}

export function clearMonadChunkEvents() {
  persistBuffer([]);
}
