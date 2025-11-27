"use client";

import { useEffect, useMemo, useState } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import type { OutboundRouterEvent } from "@/lib/api/outbound-router";
import {
  getOutboundRouterEvents,
  subscribeToOutboundRouterEvents,
} from "@/lib/storage/outbound-router-cache";
import { fetchOutboundAppMessage } from "@/lib/api/outbound-router";

function formatHash(hash: string) {
  if (!hash) return "Unmapped";
  if (hash.length <= 12) return hash;
  return `${hash.slice(0, 6)}…${hash.slice(-4)}`;
}

function formatTimestamp(
  value: string | number | Date | undefined,
  fallback = "—",
) {
  if (!value) return fallback;
  const date = typeof value === "number" ? new Date(value) : new Date(value);
  if (Number.isNaN(date.getTime())) return fallback;
  return date.toLocaleTimeString();
}

function resolveTimestamp(value: string | number | Date | undefined) {
  if (!value) return 0;
  if (typeof value === "number") return value;
  if (value instanceof Date) return value.getTime();
  const parsed = Date.parse(value);
  return Number.isNaN(parsed) ? 0 : parsed;
}

export function ChunkQueueDetailPanel() {
  const chunkQueue = useNodePulseStore((state) => state.chunkQueue);
  const [expandedHash, setExpandedHash] = useState<string | null>(null);
  const [selectedPacketId, setSelectedPacketId] = useState<string | null>(null);
  const [selectedRouterId, setSelectedRouterId] = useState<string | null>(null);
  const [routerPayloads, setRouterPayloads] = useState<Record<string, unknown>>(
    {},
  );
  const [routerErrors, setRouterErrors] = useState<Record<string, string>>({});
  const [routerLoading, setRouterLoading] = useState(false);

  const closeDetailPanel = () => setSelectedPacketId(null);
  const closeRouterPanel = () => setSelectedRouterId(null);

  const [routerEvents, setRouterEvents] = useState<OutboundRouterEvent[]>([]);

  useEffect(() => {
    setRouterEvents(getOutboundRouterEvents());
    const unsubscribe = subscribeToOutboundRouterEvents(() => {
      setRouterEvents(getOutboundRouterEvents());
    });
    return unsubscribe;
  }, []);

  const entries = useMemo(() => {
    return Object.entries(chunkQueue)
      .filter(([, packets]) => packets.length > 0)
      .sort(
        (a, b) =>
          (b[1][b[1].length - 1]?.timestamp ?? 0) -
          (a[1][a[1].length - 1]?.timestamp ?? 0),
      )
      .slice(0, 30);
  }, [chunkQueue]);

  const orderedRouterEvents = useMemo(() => {
    return [...routerEvents]
      .sort(
        (a, b) => resolveTimestamp(b.timestamp) - resolveTimestamp(a.timestamp),
      )
      .slice(0, 50);
  }, [routerEvents]);

  const selectedRouterEvent =
    orderedRouterEvents.find((event) => event._id === selectedRouterId) ?? null;

  useEffect(() => {
    if (orderedRouterEvents.length === 0) {
      setSelectedRouterId(null);
      return;
    }
    if (
      selectedRouterId &&
      !orderedRouterEvents.some((event) => event._id === selectedRouterId)
    ) {
      setSelectedRouterId(null);
    }
  }, [orderedRouterEvents, selectedRouterId]);

  const selectedPacket =
    entries
      .find(([hash]) => hash === expandedHash)?.[1]
      .find((packet) => packet.id === selectedPacketId) ?? null;

  const hydratedRouterPayload =
    selectedRouterId && routerPayloads[selectedRouterId]
      ? routerPayloads[selectedRouterId]
      : undefined;
  const routerError = selectedRouterId ? routerErrors[selectedRouterId] : null;

  useEffect(() => {
    if (!selectedRouterEvent) {
      setRouterLoading(false);
      return;
    }
    if (selectedRouterEvent.messageType !== 1) {
      setRouterLoading(false);
      return;
    }
    if (hydratedRouterPayload) {
      setRouterLoading(false);
      return;
    }
    let cancelled = false;
    setRouterLoading(true);
    fetchOutboundAppMessage(selectedRouterEvent._id)
      .then((fetchedEvent) => {
        if (cancelled) return;
        const payload = fetchedEvent.data;
        setRouterPayloads((prev) => ({
          ...prev,
          [selectedRouterEvent._id]: payload,
        }));
        setRouterErrors((prev) => {
          const next = { ...prev };
          if (!payload) {
            next[selectedRouterEvent._id] =
              "HTTP payload did not include a data field.";
          } else {
            delete next[selectedRouterEvent._id];
          }
          return next;
        });
      })
      .catch((error) => {
        if (cancelled) return;
        setRouterErrors((prev) => ({
          ...prev,
          [selectedRouterEvent._id]:
            error instanceof Error
              ? error.message
              : "Failed to load HTTP payload",
        }));
      })
      .finally(() => {
        if (!cancelled) {
          setRouterLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [selectedRouterEvent, hydratedRouterPayload]);

  const resolvedRouterPayload = selectedRouterEvent
    ? normalizeRouterPayload(
        selectedRouterEvent,
        selectedRouterEvent.messageType === 1
          ? hydratedRouterPayload
          : selectedRouterEvent.data,
      )
    : null;
  const routerLoadingActive =
    routerLoading && selectedRouterEvent?.messageType === 1;

  return (
    <section className="hud-panel chunk-detail-panel">
      <header className="panel-header">
        <div>
          <span className="text-label">Chunk Queue</span>
          <p className="text-title">All Packets</p>
        </div>
        <span className="badge" data-variant="info">
          {entries.length}
        </span>
      </header>
      <div className="chunk-panel-body">
        <div className="chunk-queue-section">
          <div className="chunk-detail-grid">
            <div className="chunk-list">
              {entries.length === 0 ? (
                <p className="chunk-placeholder">No pending packets</p>
              ) : (
                entries.map(([hash, packets]) => (
                  <button
                    key={hash}
                    type="button"
                    className={`chunk-list-item ${expandedHash === hash ? "active" : ""}`}
                    onClick={() => {
                      setExpandedHash(hash === expandedHash ? null : hash);
                      setSelectedPacketId(null);
                    }}
                  >
                    <div>
                      <span className="chunk-hash">{formatHash(hash)}</span>
                      <span className="chunk-count">{packets.length} pkts</span>
                    </div>
                    <small className="chunk-meta">
                      last {new Date(packets[packets.length - 1].timestamp).toLocaleTimeString()}
                    </small>
                  </button>
                ))
              )}
            </div>
            <div className="chunk-detail-view">
              {expandedHash === null ? (
                <p className="chunk-placeholder">Select a hash to inspect its packets.</p>
              ) : (
                <div className="chunk-packet-stack">
                  {entries
                    .find(([hash]) => hash === expandedHash)?.[1]
                    .map((packet) => (
                      <button
                        key={packet.id}
                        type="button"
                        className={`chunk-packet-chip ${selectedPacketId === packet.id ? "active" : ""}`}
                        onClick={() =>
                          setSelectedPacketId((current) =>
                            current === packet.id ? null : packet.id,
                          )
                        }
                      >
                        #{packet.chunkId}
                      </button>
                    ))}
                </div>
              )}
            </div>
          </div>
        </div>
      <div className="router-panel">
        <header className="panel-header">
          <div>
            <span className="text-label">Outbound Router</span>
            <p className="text-title">Recent Messages</p>
            </div>
            <span className="badge" data-variant="info">
              {orderedRouterEvents.length}
            </span>
        </header>
        <div className="router-grid">
          <div className="router-list">
            {orderedRouterEvents.length === 0 ? (
              <p className="chunk-placeholder">
                No outbound router payloads captured yet.
              </p>
            ) : (
              orderedRouterEvents.map((event) => (
                <button
                  key={event._id}
                  type="button"
                  className={`router-list-item ${selectedRouterEvent?._id === event._id ? "active" : ""}`}
                  onClick={() => setSelectedRouterId(event._id)}
                >
                  <div className="router-list-meta">
                    <span>{formatTimestamp(event.timestamp)}</span>
                    <span>{describeRouterEvent(event)}</span>
                  </div>
                  <span className="router-hash">
                    {formatHash(event.appMessageHash ?? "unmapped")}
                  </span>
                </button>
              ))
            )}
          </div>
        </div>
        {orderedRouterEvents.length > 0 ? (
          <p className="chunk-placeholder">
            Select a router packet to open detailed view.
          </p>
        ) : null}
      </div>
    </div>
      {selectedPacket ? (
        <div
          className="chunk-payload-modal"
          role="dialog"
          aria-live="polite"
          aria-modal="true"
        >
          <div
            className="chunk-payload-backdrop"
            onClick={closeDetailPanel}
            aria-hidden="true"
          />
          <div className="chunk-payload-panel">
            <header className="panel-header">
              <div>
                <span className="text-label">Chunk Payload</span>
                <p className="text-title">
                  #{selectedPacket.chunkId} · {formatHash(selectedPacket.appMessageHash ?? "unmapped")}
                </p>
              </div>
              <button
                type="button"
                className="chip-action"
                aria-label="Close payload panel"
                onClick={closeDetailPanel}
              >
                Close
              </button>
            </header>
            <div className="chunk-payload-meta">
              <dl>
                <div>
                  <dt>From</dt>
                  <dd>
                    {selectedPacket.fromIp}:{selectedPacket.fromPort}
                  </dd>
                </div>
                <div>
                  <dt>To</dt>
                  <dd>
                    {selectedPacket.toIp}:{selectedPacket.toPort}
                  </dd>
                </div>
                <div>
                  <dt>Size</dt>
                  <dd>{selectedPacket.size.toLocaleString()} B</dd>
                </div>
                <div>
                  <dt>Captured</dt>
                  <dd>{new Date(selectedPacket.timestamp).toLocaleTimeString()}</dd>
                </div>
              </dl>
            </div>
            <pre className="chunk-json chunk-payload-json">
              {JSON.stringify(selectedPacket.payload, null, 2)}
            </pre>
          </div>
        </div>
      ) : null}
      {selectedRouterEvent ? (
        <div
          className="router-payload-modal"
          role="dialog"
          aria-live="polite"
          aria-modal="true"
        >
          <div
            className="router-payload-backdrop"
            onClick={closeRouterPanel}
            aria-hidden="true"
          />
          <div className="router-payload-panel">
            <header className="panel-header">
              <div>
                <span className="text-label">Outbound Router Payload</span>
                <p className="text-title">
                  type {selectedRouterEvent.messageType} ·{" "}
                  {formatHash(selectedRouterEvent.appMessageHash ?? "—")}
                </p>
              </div>
              <button
                type="button"
                className="chip-action"
                aria-label="Close payload panel"
                onClick={closeRouterPanel}
              >
                Close
              </button>
            </header>
            <div className="router-meta-grid">
              <dl>
                <div>
                  <dt>Event ID</dt>
                  <dd>{selectedRouterEvent._id}</dd>
                </div>
                <div>
                  <dt>Version</dt>
                  <dd>
                    v{selectedRouterEvent.version.serializeVersion}/
                    {selectedRouterEvent.version.compressionVersion}
                  </dd>
                </div>
                <div>
                  <dt>Timestamp</dt>
                  <dd>{formatTimestamp(selectedRouterEvent.timestamp)}</dd>
                </div>
                <div>
                  <dt>App Hash</dt>
                  <dd>{formatHash(selectedRouterEvent.appMessageHash ?? "—")}</dd>
                </div>
              </dl>
            </div>
            {routerError ? (
              <p className="router-error" role="alert">
                {routerError}
              </p>
            ) : null}
            {routerLoadingActive ? (
              <p className="router-status">Fetching HTTP payload…</p>
            ) : null}
            {resolvedRouterPayload ? (
              <pre className="chunk-json router-json">
                {JSON.stringify(resolvedRouterPayload, null, 2)}
              </pre>
            ) : (
              <p className="chunk-placeholder">
                {selectedRouterEvent.messageType === 1
                  ? "Waiting for HTTP payload…"
                  : "No payload data available for this event."}
              </p>
            )}
          </div>
        </div>
      ) : null}
    </section>
  );
}

function normalizeRouterPayload(
  event: OutboundRouterEvent,
  overrideData?: unknown,
) {
  const { _id, __v, timestamp, ...rest } = event;
  return {
    ...rest,
    data: overrideData ?? event.data,
  };
}

function describeRouterEvent(event: OutboundRouterEvent) {
  const data = event.data as Record<string, any> | undefined | null;
  if (event.messageType === 1) {
    const typeId = getTypeId(data);
    if (typeId === 1) {
      const consensusLabel = getConsensusLabel(data);
      if (consensusLabel) {
        return `AppMessage - Consensus - ${consensusLabel}`;
      }
      return "AppMessage - Consensus";
    }
    if (typeId === 2) return "AppMessage - BlockSyncRequest";
    if (typeId === 3) return "AppMessage - BlockSyncResponse";
    if (typeId === 4) return "AppMessage - ForwardedTxs";
    if (typeId === 5) return "AppMessage - StateSync";
    return "AppMessage";
  }
  if (event.messageType === 2) {
    const type = getTypeId(data);
    const peerLabels: Record<number, string> = {
      1: "PeerDiscovery - Ping",
      2: "PeerDiscovery - Pong",
      3: "PeerDiscovery - PeerLookupRequest",
      4: "PeerDiscovery - PeerLookupResponse",
      5: "PeerDiscovery - FullnodeRaptorcastReq",
      6: "PeerDiscovery - FullnodeRaptorcastResp",
    };
    if (type && peerLabels[type]) {
      return peerLabels[type];
    }
    return "PeerDiscovery";
  }
  if (event.messageType === 3) {
    const type = getTypeId(data);
    const groupLabels: Record<number, string> = {
      1: "FullNodesGroup - PrepareRequest",
      2: "FullNodesGroup - PrepareResponse",
      3: "FullNodesGroup - ConfirmGroup",
    };
    if (type && groupLabels[type]) {
      return groupLabels[type];
    }
    return "FullNodesGroup";
  }
  return `type ${event.messageType}`;
}

function getConsensusLabel(data: Record<string, any> | null | undefined) {
  if (!data) return null;
  const stageOne = data.payload;
  const stageTwo = stageOne?.payload;
  const messageType =
    typeof stageTwo?.messageType === "number"
      ? stageTwo.messageType
      : Number(stageTwo?.messageType);
  const labels: Record<number, string> = {
    1: "Proposal",
    2: "Vote",
    3: "Timeout",
    4: "RoundRecovery",
    5: "NoEndorsement",
    6: "AdvancedRound",
  };
  if (Number.isFinite(messageType) && labels[messageType]) {
    return labels[messageType];
  }
  return null;
}

function getTypeId(value: Record<string, any> | undefined | null) {
  if (!value) return null;
  const raw = value.typeId ?? value.TypeID ?? value.type ?? value.Type;
  if (raw === undefined) return null;
  const num = typeof raw === "number" ? raw : Number(raw);
  return Number.isFinite(num) ? num : null;
}
