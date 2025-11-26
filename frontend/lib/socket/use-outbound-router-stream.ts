"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import {
  appendOutboundRouterEvent,
  clearOutboundRouterEvents,
} from "@/lib/storage/outbound-router-cache";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "OUTBOUND_ROUTER";

export function useOutboundRouterStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;

  useEffect(() => {
    if (!endpoint) {
      return;
    }

    clearOutboundRouterEvents();
    const socket = acquireSocket(endpoint);

    const handleConnect = () => {
      clearOutboundRouterEvents();
    };

    const handlePayload = async (payload: unknown) => {
      try {
        const result = await appendOutboundRouterEvent(payload);
        const hash =
          result.event.appMessageHash ??
          (result.event.data &&
          typeof result.event.data === "object" &&
          "appMessageHash" in result.event.data
            ? (result.event.data as { appMessageHash?: string })
                .appMessageHash
            : undefined);
        if (hash) {
          useNodePulseStore.getState().triggerChunkAssembly(hash);
        }
      } catch {
        // parsing/hydration errors are already logged
      }
    };

    socket.on("connect", handleConnect);
    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off("connect", handleConnect);
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint]);
}
