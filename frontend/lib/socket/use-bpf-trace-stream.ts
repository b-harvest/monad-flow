"use client";

import { useEffect } from "react";
import { appendBpfTraceEvent } from "@/lib/storage/bpf-trace-cache";
import { defaultSocketEndpoint } from "./config";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "BPF_TRACE";

/**
 * Subscribes to the backend Socket.IO bridge and caches BPF trace events.
 */
export function useBpfTraceStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;

  useEffect(() => {
    if (!endpoint) {
      return;
    }

    const socket = acquireSocket(endpoint);

    const handlePayload = (payload: unknown) => {
      try {
        appendBpfTraceEvent(payload);
      } catch {
        // Parsing errors are surfaced inside appendBpfTraceEvent.
      }
    };

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint]);
}
