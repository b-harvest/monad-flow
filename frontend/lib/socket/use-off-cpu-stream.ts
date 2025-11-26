"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import { appendOffCpuEvent } from "@/lib/storage/off-cpu-cache";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "OFF_CPU";

export function useOffCpuStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;

  useEffect(() => {
    if (!endpoint) {
      return;
    }

    const socket = acquireSocket(endpoint);

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
      releaseSocket();
    };
  }, [endpoint]);
}
