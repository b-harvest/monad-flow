"use client";

import { useEffect, useRef } from "react";
import { defaultSocketEndpoint } from "./config";
import { MonadChunkParser } from "@/lib/api/monad-chunk";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { prepareChunkData } from "@/lib/monad/chunk-event-handler";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "MONAD_CHUNK";

export function useMonadChunkStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;
  const playbackMode = useNodePulseStore((state) => state.playback.mode);
  const bufferRef = useRef<any[]>([]);

  useEffect(() => {
    if (playbackMode !== "live" || !endpoint) {
      return;
    }

    const socket = acquireSocket(endpoint);

    const handleConnect = () => {
      const state = useNodePulseStore.getState();
      state.setConnectionStatus("connected");
      state.ensureLocalNode();
    };

    const handleDisconnect = () => {
      const state = useNodePulseStore.getState();
      state.setConnectionStatus("lost");
      state.resetNetworkGraph();
    };

    const handlePayload = (payload: unknown) => {
      // Buffer raw payloads
      bufferRef.current.push(payload);
    };

    socket.on("connect", handleConnect);
    socket.on("disconnect", handleDisconnect);
    socket.on(EVENT_NAME, handlePayload);

    const flushInterval = setInterval(() => {
      if (bufferRef.current.length === 0) return;

      const batch = bufferRef.current.splice(0);
      const state = useNodePulseStore.getState();
      const localIp = state.localNodeIp;
      const localId = state.ensureLocalNode(); // Ensure local node exists and get ID

      const preparedItems = batch
        .map((payload) => {
          const result = MonadChunkParser.safeParse(payload);
          if (!result.success) {
            console.error("[MONAD_CHUNK] Failed to parse payload", result.error);
            return null;
          }
          return prepareChunkData(result.data, localIp, localId);
        })
        .filter((item) => item !== null);

      if (preparedItems.length > 0) {
        state.batchIngestChunks(preparedItems);
      }
    }, 100);

    const pruneInterval = window.setInterval(() => {
      useNodePulseStore.getState().pruneInactiveNodes();
    }, 10000);

    return () => {
      socket.off("connect", handleConnect);
      socket.off("disconnect", handleDisconnect);
      socket.off(EVENT_NAME, handlePayload);
      clearInterval(flushInterval);
      window.clearInterval(pruneInterval);
      releaseSocket();
    };
  }, [endpoint, playbackMode]);
}
