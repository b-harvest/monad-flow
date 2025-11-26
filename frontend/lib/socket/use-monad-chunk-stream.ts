"use client";

import { useEffect } from "react";
import { defaultSocketEndpoint } from "./config";
import { appendMonadChunkEvent } from "@/lib/storage/monad-chunk-cache";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { acquireSocket, releaseSocket } from "./shared-socket";

const EVENT_NAME = "MONAD_CHUNK";

export function useMonadChunkStream(options?: { endpoint?: string }) {
  const endpoint = options?.endpoint ?? defaultSocketEndpoint;

  useEffect(() => {
    if (!endpoint) {
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
      try {
        const { event } = appendMonadChunkEvent(payload);
        const state = useNodePulseStore.getState();
        const localId = state.ensureLocalNode();
        const localIp = state.localNodeIp;
        const srcIp = event.network.ipv4.srcIp;
        const dstIp = event.network.ipv4.dstIp;
        const srcPort = event.network.port.srcPort;
        const dstPort = event.network.port.dstPort;

        let direction: "inbound" | "outbound" = "inbound";
        let peerIp = srcIp;
        let peerPort = srcPort;
        const hasLocal =
          localIp && localIp !== "64.31.48.109" && localIp.trim().length > 0;

        if (hasLocal && srcIp === localIp) {
          direction = "outbound";
          peerIp = dstIp;
          peerPort = dstPort;
        } else if (hasLocal && dstIp === localIp) {
          direction = "inbound";
          peerIp = srcIp;
          peerPort = srcPort;
        }

        const peerId = state.upsertChunkPeer(peerIp, peerPort);
        state.addEffect({
          id: `chunk-${event.chunkId}-${event.timestamp}-${Math.random().toString(36).slice(2, 6)}`,
          type: "chunk",
          fromNodeId: direction === "inbound" ? peerId : localId,
          toNodeId: direction === "inbound" ? localId : peerId,
          createdAt: Date.now(),
          ttl: 1500,
          direction,
        });
        state.pushChunkPacket({
          id: `packet-${event.chunkId}-${event.timestamp}-${Math.random().toString(36).slice(2, 6)}`,
          appMessageHash: event.appMessageHash,
          chunkId: event.chunkId,
          timestamp: Date.now(),
          fromIp: direction === "inbound" ? peerIp : localIp,
          fromPort: direction === "inbound" ? peerPort : event.network.port.srcPort,
          toIp: direction === "inbound" ? localIp : peerIp,
          toPort: direction === "inbound" ? event.network.port.dstPort : peerPort,
          size: event.appMessageLen ?? event.merkleProof?.length ?? 0,
          payload: event,
        });
      } catch {
        // parsing errors already logged in cache helper
      }
    };

    socket.on("connect", handleConnect);
    socket.on("disconnect", handleDisconnect);
    socket.on(EVENT_NAME, handlePayload);

    const pruneInterval = window.setInterval(() => {
      useNodePulseStore.getState().pruneInactiveNodes();
    }, 10000);

    return () => {
      socket.off("connect", handleConnect);
      socket.off("disconnect", handleDisconnect);
      socket.off(EVENT_NAME, handlePayload);
      window.clearInterval(pruneInterval);
      releaseSocket();
    };
  }, [endpoint]);
}
