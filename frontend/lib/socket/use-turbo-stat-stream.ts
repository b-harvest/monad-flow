"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import { TurboStatParser } from "@/lib/api/turbo-stat";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { acquireSocket, releaseSocket } from "./shared-socket";
import { createBufferedHandler } from "./buffered-handler";

const EVENT_NAME = "TURBO_STAT";

export function useTurboStatStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;
  const playbackMode = useNodePulseStore((state) => state.playback.mode);

  useEffect(() => {
    if (playbackMode !== "live" || !endpoint) {
      return;
    }
    const socket = acquireSocket(endpoint);

    const handlePayload = createBufferedHandler((payload: unknown) => {
      const result = TurboStatParser.safeParse(payload);
      if (!result.success) {
        console.error("[TURBO_STAT] Failed to parse payload", result.error);
        return;
      }
      useNodePulseStore.getState().pushTurboStatEvent(result.data);
    });

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint, playbackMode]);
}
