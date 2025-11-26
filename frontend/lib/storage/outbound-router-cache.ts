"use client";

import { z } from "zod";
import type { OutboundRouterEvent } from "@/lib/api/outbound-router";
import {
  OutboundRouterEventSchema,
  fetchOutboundAppMessage,
} from "@/lib/api/outbound-router";

const STORAGE_KEY = "monad-flow:outbound-router";
export const MAX_OUTBOUND_ROUTER_ENTRIES = 50;

const EventArraySchema = z.array(OutboundRouterEventSchema);
type Listener = () => void;
const listeners = new Set<Listener>();

let hydratedBuffer: OutboundRouterEvent[] = [];
let isHydrated = false;

function notifyListeners() {
  listeners.forEach((listener) => {
    try {
      listener();
    } catch (error) {
      console.warn("[OUTBOUND_ROUTER] Listener error:", error);
    }
  });
}

function loadFromStorage(): OutboundRouterEvent[] {
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

function persistBuffer(next: OutboundRouterEvent[]) {
  hydratedBuffer = next;
  if (typeof window !== "undefined") {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch (error) {
      console.warn("[OUTBOUND_ROUTER] Failed to persist cache:", error);
    }
  }
  notifyListeners();
}

export function getOutboundRouterEvents() {
  return loadFromStorage();
}

export function subscribeToOutboundRouterEvents(listener: Listener) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export async function appendOutboundRouterEvent(payload: unknown) {
  const parsed = OutboundRouterEventSchema.safeParse(payload);
  if (!parsed.success) {
    console.error("[OUTBOUND_ROUTER] Parse error:", parsed.error, payload);
    throw parsed.error;
  }

  let event = parsed.data;

  const needsHydration =
    event.messageType === 1 && (event.data === undefined || event.data === null);

  if (needsHydration) {
    try {
      const fetchedEvent = await fetchOutboundAppMessage(event._id);
      event = {
        ...event,
        ...fetchedEvent,
        type: event.type ?? fetchedEvent.type,
        data: fetchedEvent.data,
      };
    } catch (error) {
      console.warn("[OUTBOUND_ROUTER] Failed to hydrate payload via HTTP:", error);
    }
  }

  const current = loadFromStorage();
  const next = [...current, event].slice(-MAX_OUTBOUND_ROUTER_ENTRIES);
  persistBuffer(next);
  return { event, snapshot: next };
}

export function clearOutboundRouterEvents() {
  persistBuffer([]);
}
