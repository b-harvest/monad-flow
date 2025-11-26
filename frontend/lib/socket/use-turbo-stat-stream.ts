"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import { appendTurboStatEvent } from "@/lib/storage/turbo-stat-cache";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "TURBO_STAT";

export function useTurboStatStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;

  useEffect(() => {
    if (!endpoint) {
      return;
    }
    const socket = acquireSocket(endpoint);

    const handlePayload = (payload: unknown) => {
      try {
        appendTurboStatEvent(payload);
      } catch {
        // parse errors already logged in cache helper
      }
    };

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint]);
}
