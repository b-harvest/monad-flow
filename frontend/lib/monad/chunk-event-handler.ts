"use client";

import type { MonadChunkEvent } from "@/lib/api/monad-chunk";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";

const DEFAULT_DST_PORT = 0;

export interface PreparedChunkData {
  peer: { ip: string; port: number };
  effect: any; // Using any to avoid circular dependency with types/monad if not imported, but better to import PulseVisualEffect
  packet: any; // Same for ChunkPacketRecord
}

import type { PulseVisualEffect, ChunkPacketRecord } from "@/types/monad";

export function prepareChunkData(
  event: MonadChunkEvent,
  localIp: string,
  localId: string,
): PreparedChunkData {
  const srcIp = event.network.ipv4.srcIp;
  const dstIp = event.network.ipv4.dstIp;
  const srcPort = event.network.port.srcPort ?? DEFAULT_DST_PORT;
  const dstPort = event.network.port.dstPort ?? DEFAULT_DST_PORT;

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

  const effect: PulseVisualEffect = {
    id: `chunk-${event.chunkId}-${event.timestamp}-${Math.random()
      .toString(36)
      .slice(2, 6)}`,
    type: "chunk",
    fromNodeId: direction === "inbound" ? `chunk-${peerIp}:${peerPort}` : localId,
    toNodeId: direction === "inbound" ? localId : `chunk-${peerIp}:${peerPort}`,
    createdAt: Date.now(),
    ttl: 1500,
    direction,
  };

  const packet: ChunkPacketRecord = {
    id: `packet-${event.chunkId}-${event.timestamp}-${Math.random()
      .toString(36)
      .slice(2, 6)}`,
    appMessageHash: event.appMessageHash,
    chunkId: event.chunkId,
    timestamp: Date.now(),
    fromIp: direction === "inbound" ? peerIp : localIp,
    fromPort:
      direction === "inbound"
        ? peerPort
        : event.network.port.srcPort ?? DEFAULT_DST_PORT,
    toIp: direction === "inbound" ? localIp : peerIp,
    toPort:
      direction === "inbound"
        ? event.network.port.dstPort ?? DEFAULT_DST_PORT
        : peerPort,
    size: event.appMessageLen ?? event.merkleProof?.length ?? 0,
    payload: event,
  };

  return {
    peer: { ip: peerIp, port: peerPort },
    effect,
    packet,
  };
}
