"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import { appendSchedulerEvent } from "@/lib/storage/scheduler-cache";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "SCHEDULER";

export function useSchedulerStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;

  useEffect(() => {
    if (!endpoint) {
      return;
    }

    const socket = acquireSocket(endpoint);

    const handlePayload = (payload: unknown) => {
      try {
        appendSchedulerEvent(payload);
      } catch {
        // validation errors already logged
      }
    };

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint]);
}
