"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import { LeaderEventSchema } from "@/lib/api/leader";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "LEADER";

export function useLeaderStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;
  const playbackMode = useNodePulseStore((state) => state.playback.mode);

  useEffect(() => {
    if (playbackMode !== "live" || !endpoint) {
      return;
    }

    const socket = acquireSocket(endpoint);

    const handlePayload = (payload: unknown) => {
      const result = LeaderEventSchema.safeParse(payload);
      if (!result.success) {
        console.error("[LEADER] Failed to parse payload", result.error);
        return;
      }
      useNodePulseStore.getState().addLeader(result.data);
    };

    socket.on(EVENT_NAME, handlePayload);

    return () => {
      socket.off(EVENT_NAME, handlePayload);
      releaseSocket();
    };
  }, [endpoint, playbackMode]);
}
