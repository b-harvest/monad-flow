"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import { appendPerfStatEvent } from "@/lib/storage/perf-stat-cache";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "PERF_STAT";

export function usePerfStatStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;

  useEffect(() => {
    if (!endpoint) {
      return;
    }
    const socket = acquireSocket(endpoint);

    const handlePayload = (payload: unknown) => {
      try {
        appendPerfStatEvent(payload);
      } catch {
        // parse errors already logged
      }
    };

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint]);
}
