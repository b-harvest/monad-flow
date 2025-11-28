"use client";

import { useEffect } from "react";
import { BpfTraceEventSchema } from "@/lib/api/bpf-trace";
import { defaultSocketEndpoint } from "./config";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { acquireSocket, releaseSocket } from "./shared-socket";
import { createBufferedHandler } from "./buffered-handler";

const EVENT_NAME = "BPF_TRACE";

/**
 * Subscribes to the backend Socket.IO bridge and caches BPF trace events.
 */
export function useBpfTraceStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;
  const playbackMode = useNodePulseStore((state) => state.playback.mode);

  useEffect(() => {
    if (playbackMode !== "live" || !endpoint) {
      return;
    }

    const socket = acquireSocket(endpoint);

    const handlePayload = createBufferedHandler((payload: unknown) => {
      const result = BpfTraceEventSchema.safeParse(payload);
      if (!result.success) {
        console.error("[BPF_TRACE] Failed to parse payload", result.error);
        return;
      }
      useNodePulseStore.getState().pushBpfTraceEvent(result.data);
    });

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint, playbackMode]);
}
