"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import {
  OutboundRouterEventSchema,
  type OutboundRouterEvent,
} from "@/lib/api/outbound-router";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { hydrateOutboundRouterEvent } from "@/lib/monad/hydrate-outbound-router-event";

import { acquireSocket, releaseSocket } from "./shared-socket";
import { createBufferedHandler } from "./buffered-handler";

const EVENT_NAME = "OUTBOUND_ROUTER";

export function useOutboundRouterStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;
  const playbackMode = useNodePulseStore((state) => state.playback.mode);

  useEffect(() => {
    if (playbackMode !== "live" || !endpoint) {
      return;
    }

    useNodePulseStore.getState().clearRouterEvents();
    const socket = acquireSocket(endpoint);

    const handleConnect = () => {
      useNodePulseStore.getState().clearRouterEvents();
    };

    const handlePayload = createBufferedHandler((payload: unknown) => {
      const result = OutboundRouterEventSchema.safeParse(payload);
      if (!result.success) {
        console.error(
          "[OUTBOUND_ROUTER] Failed to parse payload",
          result.error,
        );
        return;
      }
      hydrateOutboundRouterEvent(result.data)
        .then(processRouterEvent)
        .catch(() => processRouterEvent(result.data));
    });

    socket.on("connect", handleConnect);
    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off("connect", handleConnect);
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint, playbackMode]);
}

function processRouterEvent(event: OutboundRouterEvent) {
  const state = useNodePulseStore.getState();
  state.pushRouterEvent(event);
  const hash = getRouterEventHash(event);
  if (hash) {
    state.triggerChunkAssembly(hash);
  }
}

function getRouterEventHash(event: OutboundRouterEvent) {
  if (typeof event.appMessageHash === "string") {
    return event.appMessageHash;
  }
  if (
    event.data &&
    typeof event.data === "object" &&
    "appMessageHash" in event.data &&
    typeof (event.data as { appMessageHash?: string }).appMessageHash === "string"
  ) {
    return (event.data as { appMessageHash?: string }).appMessageHash;
  }
  return undefined;
}
