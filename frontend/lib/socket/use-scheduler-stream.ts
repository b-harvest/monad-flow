"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import { SchedulerParser } from "@/lib/api/scheduler";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { acquireSocket, releaseSocket } from "./shared-socket";
import { createBufferedHandler } from "./buffered-handler";

const EVENT_NAME = "SCHEDULER";

export function useSchedulerStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;
  const playbackMode = useNodePulseStore((state) => state.playback.mode);

  useEffect(() => {
    if (playbackMode !== "live" || !endpoint) {
      return;
    }

    const socket = acquireSocket(endpoint);

    const handlePayload = createBufferedHandler((payload: unknown) => {
      const result = SchedulerParser.safeParse(payload);
      if (!result.success) {
        console.error("[SCHEDULER] Failed to parse payload", result.error);
        return;
      }
      useNodePulseStore.getState().pushSchedulerEvent(result.data);
    });

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint, playbackMode]);
}
