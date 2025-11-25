"use client";

import { useEffect } from "react";
import { io, Socket } from "socket.io-client";
import { appendSystemLogEvent } from "@/lib/storage/system-log-cache";
import { defaultSocketEndpoint } from "./config";

const EVENT_NAME = "SYSTEM_LOG";

export function useSystemLogStream(options?: { endpoint?: string }) {
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
        appendSystemLogEvent(payload);
      } catch (error) {
        console.error("[SYSTEM_LOG] Failed to cache event", error);
      }
    };

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      socket.disconnect();
    };
  }, [endpoint]);
}
