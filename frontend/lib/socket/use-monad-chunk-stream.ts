"use client";

import { useEffect } from "react";
import { io, Socket } from "socket.io-client";
import { defaultSocketEndpoint } from "./config";
import { appendMonadChunkEvent } from "@/lib/storage/monad-chunk-cache";

const EVENT_NAME = "MONAD_CHUNK";

export function useMonadChunkStream(options?: { endpoint?: string }) {
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
        appendMonadChunkEvent(payload);
      } catch {
        // parsing errors already logged in cache helper
      }
    };

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      socket.disconnect();
    };
  }, [endpoint]);
}
