"use client";

import { useEffect } from "react";
import { SystemLogEventSchema } from "@/lib/api/system-log";
import { defaultSocketEndpoint } from "./config";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { acquireSocket, releaseSocket } from "./shared-socket";
import { createBufferedHandler } from "./buffered-handler";

const EVENT_NAME = "SYSTEM_LOG";

export function useSystemLogStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;
  const playbackMode = useNodePulseStore((state) => state.playback.mode);

  useEffect(() => {
    if (playbackMode !== "live" || !endpoint) {
      return;
    }

    const socket = acquireSocket(endpoint);

    const handlePayload = createBufferedHandler((payload: unknown) => {
      const result = SystemLogEventSchema.safeParse(payload);
      if (!result.success) {
        console.error("[SYSTEM_LOG] Failed to parse event", result.error);
        return;
      }
      useNodePulseStore.getState().pushSystemLogEvent(result.data);
    });

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint, playbackMode]);
}
