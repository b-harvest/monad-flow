"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import { PerfStatParser } from "@/lib/api/perf-stat";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { acquireSocket, releaseSocket } from "./shared-socket";
import { createBufferedHandler } from "./buffered-handler";

const EVENT_NAME = "PERF_STAT";

export function usePerfStatStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;
  const playbackMode = useNodePulseStore((state) => state.playback.mode);

  useEffect(() => {
    if (playbackMode !== "live" || !endpoint) {
      return;
    }
    const socket = acquireSocket(endpoint);

    const handlePayload = createBufferedHandler((payload: unknown) => {
      const result = PerfStatParser.safeParse(payload);
      if (!result.success) {
        console.error("[PERF_STAT] Failed to parse payload", result.error);
        return;
      }
      useNodePulseStore.getState().pushPerfStatEvent(result.data);
    });

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint, playbackMode]);
}
