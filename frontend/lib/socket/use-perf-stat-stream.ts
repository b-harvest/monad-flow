"use client";

import { useEffect } from "react";
import { io, Socket } from "socket.io-client";
import { defaultSocketEndpoint } from "./config";
import { appendPerfStatEvent } from "@/lib/storage/perf-stat-cache";

const EVENT_NAME = "PERF_STAT";

export function usePerfStatStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;

  useEffect(() => {
    if (!endpoint) {
      return;
    }
    const socket: Socket = io(endpoint, {
      transports: ["websocket", "polling"],
    });

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
      socket.disconnect();
    };
  }, [endpoint]);
}
