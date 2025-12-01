"use client";

import { useEffect, useRef } from "react";
import { defaultSocketEndpoint } from "./config";
import { PingLatencySchema } from "@/lib/api/ping-latency";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "PING";

export function usePingLatencyStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;
  const playbackMode = useNodePulseStore((state) => state.playback.mode);
  const bufferRef = useRef<any[]>([]);

  useEffect(() => {
    if (playbackMode !== "live" || !endpoint) {
      return;
    }

    const socket = acquireSocket(endpoint);

    const handlePayload = (payload: unknown) => {
      bufferRef.current.push(payload);
    };

    socket.on(EVENT_NAME, handlePayload);

    const flushInterval = setInterval(() => {
      if (bufferRef.current.length === 0) return;

      const batch = bufferRef.current.splice(0);
      const state = useNodePulseStore.getState();

      const preparedItems = batch
        .map((payload) => {
          const result = PingLatencySchema.safeParse(payload);
          if (!result.success) {
            console.error("[PING] Failed to parse payload", result.error);
            return null;
          }
          return result.data;
        })
        .filter((item) => item !== null && item.rtt_ms !== undefined) as {
        ip: string;
        rtt_ms: number;
      }[];

      if (preparedItems.length > 0) {
        state.batchIngestPings(preparedItems);
      }
    }, 1000); // Update latency every second to avoid too many re-renders

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      clearInterval(flushInterval);
      releaseSocket();
    };
  }, [endpoint, playbackMode]);
}
