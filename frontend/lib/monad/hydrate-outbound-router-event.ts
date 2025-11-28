import type { OutboundRouterEvent } from "@/lib/api/outbound-router";
import { fetchOutboundAppMessage } from "@/lib/api/outbound-router";

function needsHydration(event: OutboundRouterEvent) {
  if (event.messageType !== 1) {
    return false;
  }
  if (event.data === null || event.data === undefined) {
    return true;
  }
  if (typeof event.data !== "object") {
    return false;
  }
  return Object.keys(event.data as Record<string, unknown>).length === 0;
}

export async function hydrateOutboundRouterEvent(
  event: OutboundRouterEvent,
): Promise<OutboundRouterEvent> {
  if (!needsHydration(event) || !event._id) {
    return event;
  }
  try {
    const fetched = await fetchOutboundAppMessage(event._id);
    return {
      ...event,
      ...fetched,
      type: event.type ?? fetched.type,
      version: event.version ?? fetched.version,
      data: fetched.data ?? event.data,
    };
  } catch (error) {
    console.warn("[OUTBOUND_ROUTER] Failed to hydrate via HTTP", error);
    return event;
  }
}
