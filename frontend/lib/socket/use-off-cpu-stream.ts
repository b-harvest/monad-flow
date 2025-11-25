"use client";

import { useEffect } from "react";
import { io, Socket } from "socket.io-client";
import { defaultSocketEndpoint } from "./config";
import { appendOffCpuEvent } from "@/lib/storage/off-cpu-cache";

const EVENT_NAME = "OFF_CPU";

export function useOffCpuStream(options?: { endpoint?: string }) {
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
        appendOffCpuEvent(payload);
      } catch {
        // validation already logs details
      }
    };

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      socket.disconnect();
    };
  }, [endpoint]);
}
