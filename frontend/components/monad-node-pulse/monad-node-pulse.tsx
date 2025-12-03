"use client";

import {
  Suspense,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";
import { NodePulseMap } from "./node-pulse-map";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { MetricsPanel } from "./panels/metrics-panel";
import { ChunkQueueDetailPanel } from "./panels/chunk-queue-detail-panel";
import { LeaderSchedulePanel } from "./panels/leader-schedule-panel";
import { NodeTelemetryPanel } from "./panels/node-telemetry-panel";
import { PidTelemetryPanel } from "./panels/pid-telemetry-panel";
import { SystemLogPanel } from "./panels/system-log-panel";
import { TransactionFeedPanel } from "./panels/transaction-feed-panel";
import { PingLatencyPanel } from "./panels/ping-latency-panel";

import { CommandNav } from "./top-nav";
import type { MonadChunkEvent } from "@/lib/api/monad-chunk";
import type { OutboundRouterEvent } from "@/lib/api/outbound-router";
import type { PingLatencyEvent } from "@/lib/api/ping-latency";
import type { BpfTraceEvent } from "@/lib/api/bpf-trace";
import type { SystemLogEvent } from "@/lib/api/system-log";
import type { OffCpuEvent } from "@/lib/api/off-cpu";
import type { SchedulerEvent } from "@/lib/api/scheduler";
import type { PerfStatEvent } from "@/lib/api/perf-stat";
import type { TurboStatEvent } from "@/lib/api/turbo-stat";
import type { HistoricalEvent, PlaybackState } from "@/types/monad";
import { prepareChunkData } from "@/lib/monad/chunk-event-handler";
import { hydrateOutboundRouterEvent } from "@/lib/monad/hydrate-outbound-router-event";
import { useBpfTraceStream } from "@/lib/socket/use-bpf-trace-stream";
import { useSystemLogStream } from "@/lib/socket/use-system-log-stream";
import { usePerfStatStream } from "@/lib/socket/use-perf-stat-stream";
import { useOffCpuStream } from "@/lib/socket/use-off-cpu-stream";
import { useSchedulerStream } from "@/lib/socket/use-scheduler-stream";
import { useTurboStatStream } from "@/lib/socket/use-turbo-stat-stream";
import { useMonadChunkStream } from "@/lib/socket/use-monad-chunk-stream";
import { usePingLatencyStream } from "@/lib/socket/use-ping-latency-stream";
import { useOutboundRouterStream } from "@/lib/socket/use-outbound-router-stream";
import { useLeaderStream } from "@/lib/socket/use-leader-stream";

const MonadNodePulse = () => {
  useBpfTraceStream();
  useSystemLogStream();
  useOffCpuStream();
  usePerfStatStream();
  useSchedulerStream();
  useTurboStatStream();
  useMonadChunkStream();
  usePingLatencyStream();
  usePingLatencyStream();
  useOutboundRouterStream();
  useLeaderStream();
  const metrics = useNodePulseStore((state) => state.metrics);
  const nodes = useNodePulseStore((state) => state.nodes);
  const playback = useNodePulseStore((state) => state.playback);
  const setPlayback = useNodePulseStore((state) => state.setPlayback);
  const historicalTimeline = useNodePulseStore(
    (state) => state.historicalTimeline,
  );
  const historicalPointer = useNodePulseStore(
    (state) => state.historicalPointer,
  );
  const setHistoricalPointer = useNodePulseStore(
    (state) => state.setHistoricalPointer,
  );
  const playbackRef = useRef(playback);

  const [processTelemetryVisible, setProcessTelemetryVisible] = useState(false);
  const [mapShellEl, setMapShellEl] = useState<HTMLDivElement | null>(null);
  const mapShellRef = useCallback((node: HTMLDivElement | null) => {
    setMapShellEl(node);
  }, []);
  const [mapShellBounds, setMapShellBounds] = useState<DOMRect | null>(null);
  const [hasMounted, setHasMounted] = useState(false);
  const [historicalLoading, setHistoricalLoading] = useState(false);
  const [historicalError, setHistoricalError] = useState<string | null>(null);
  const historicalControllerRef = useRef<AbortController | null>(null);
  const MAX_HISTORICAL_WINDOW_MS = 5 * 60 * 1000;

  useEffect(() => {
    setHasMounted(true);
  }, []);

  useEffect(() => {
    if (!mapShellEl) return;
    const measure = () => {
      setMapShellBounds(mapShellEl.getBoundingClientRect());
    };
    measure();
    if (!processTelemetryVisible) {
      return;
    }
    let observer: ResizeObserver | null = null;
    if (typeof ResizeObserver !== "undefined") {
      observer = new ResizeObserver(measure);
      observer.observe(mapShellEl);
    }
    window.addEventListener("resize", measure);
    window.addEventListener("scroll", measure, true);
    return () => {
      observer?.disconnect();
      window.removeEventListener("resize", measure);
      window.removeEventListener("scroll", measure, true);
    };
  }, [mapShellEl, processTelemetryVisible]);

  useEffect(() => {
    playbackRef.current = playback;
  }, [playback]);

  useEffect(() => {
    const state = useNodePulseStore.getState();
    if (playback.mode === "historical") {
      state.setConnectionStatus("degraded");
      state.resetNetworkGraph();
    } else {
      state.setConnectionStatus("connected");
      state.resetHistoricalTimeline();
    }
  }, [playback.mode]);

  const handleHistoricalFetch = async (range: { from: number; to: number }) => {
    historicalControllerRef.current?.abort();
    const controller = new AbortController();
    historicalControllerRef.current = controller;
    const from = range.from;
    const to = range.to;
    if (!Number.isFinite(from) || !Number.isFinite(to) || to <= from) {
      setHistoricalError("Start time must be before end time.");
      return;
    }
    if (to - from > MAX_HISTORICAL_WINDOW_MS) {
      setHistoricalError("Selected range is too large. Please choose 5 minutes or less.");
      return;
    }
    setHistoricalError(null);
    setHistoricalLoading(true);
    const state = useNodePulseStore.getState();
    state.resetNetworkGraph();
    state.clearRouterEvents();
    state.setConnectionStatus("degraded");

    try {
      const timeline = await loadHistoricalLogs({ from, to }, controller);
      useNodePulseStore.getState().setHistoricalTimeline(timeline);
      setHistoricalError(null);
      setPlayback({
        ...playback,
        cursor: from,
        range: { from, to },
        mode: "historical",
        isPlaying: false,
      });
    } catch (error) {
      if ((error as Error).name === "AbortError") {
        return;
      }
      setHistoricalError(
        error instanceof Error
          ? error.message
          : "Failed to load historical logs.",
      );
    } finally {
      if (historicalControllerRef.current === controller) {
        setHistoricalLoading(false);
      }
    }
  };

  useEffect(() => {
    if (playback.mode !== "historical") {
      return;
    }
    let nextIndex = historicalPointer;
    const cursorTime = playback.cursor;
    while (
      nextIndex < historicalTimeline.length &&
      historicalTimeline[nextIndex].timestamp <= cursorTime
    ) {
      const entry = historicalTimeline[nextIndex];
      if (entry.type === "chunk") {
        const state = useNodePulseStore.getState();
        const localIp = state.localNodeIp;
        const localId = state.ensureLocalNode();
        const prepared = prepareChunkData(
          entry.payload as MonadChunkEvent,
          localIp,
          localId,
        );
        state.batchIngestChunks([prepared]);
      } else if (entry.type === "router") {
        hydrateOutboundRouterEvent(entry.payload)
          .then((hydrated) => ingestRouterEvent(hydrated))
          .catch(() => ingestRouterEvent(entry.payload));
      } else if (entry.type === "ping") {
        const state = useNodePulseStore.getState();
        const payload = entry.payload as PingLatencyEvent;
        if (payload.ip && typeof payload.rtt_ms === "number") {
          state.batchIngestPings([
            { ip: payload.ip, rtt_ms: payload.rtt_ms },
          ]);
        }
      } else if (entry.type === "bpf") {
        const state = useNodePulseStore.getState();
        state.pushBpfTraceEvent(entry.payload as BpfTraceEvent);
      } else if (entry.type === "system") {
        const state = useNodePulseStore.getState();
        state.pushSystemLogEvent(entry.payload as SystemLogEvent);
      } else if (entry.type === "offcpu") {
        const state = useNodePulseStore.getState();
        state.pushOffCpuEvent(entry.payload as OffCpuEvent);
      } else if (entry.type === "scheduler") {
        const state = useNodePulseStore.getState();
        state.pushSchedulerEvent(entry.payload as SchedulerEvent);
      } else if (entry.type === "perf") {
        const state = useNodePulseStore.getState();
        state.pushPerfStatEvent(entry.payload as PerfStatEvent);
      } else if (entry.type === "turbo") {
        const state = useNodePulseStore.getState();
        state.pushTurboStatEvent(entry.payload as TurboStatEvent);
      }
      nextIndex += 1;
    }
    if (nextIndex !== historicalPointer) {
      setHistoricalPointer(nextIndex);
    }
  }, [
    playback.mode,
    playback.cursor,
    historicalTimeline,
    historicalPointer,
    setHistoricalPointer,
  ]);

  useEffect(() => {
    if (playback.mode !== "historical" || !playback.isPlaying) {
      return;
    }
    let raf: number;
    let last = performance.now();
    const step = () => {
      const now = performance.now();
      const delta = now - last;
      last = now;
      const state = useNodePulseStore.getState();
      const current = state.playback;
      if (current.mode !== "historical" || !current.isPlaying) {
        return;
      }
      const nextCursor = Math.min(
        current.range.to,
        current.cursor + delta * current.speed,
      );
      const nextState: Partial<PlaybackState> = {
        cursor: nextCursor,
        isPlaying: nextCursor < current.range.to,
      };
      state.setPlayback(nextState);
      raf = window.requestAnimationFrame(step);
    };
    raf = window.requestAnimationFrame(step);
    return () => {
      window.cancelAnimationFrame(raf);
    };
  }, [playback.mode, playback.isPlaying]);

  return (
    <div className="node-pulse-shell">
      <CommandNav
        connectionStatus={metrics.connectionStatus}
        playback={playback}
        onPlaybackChange={setPlayback}
        historicalLoading={historicalLoading}
        historicalError={historicalError}
        onHistoricalFetch={handleHistoricalFetch}
      />
      <div className="node-pulse-layout">
        <section className="node-pulse-stage">
          <div className="node-pulse-stage-grid">
            <NodeTelemetryPanel />
            <div className="node-pulse-map-column">
              <div className="node-pulse-map-shell" ref={mapShellRef}>
                <Suspense fallback={<div className="visualization-map" />}>
                  <NodePulseMap />
                </Suspense>
              </div>
              <div className="metrics-row">
                <div>
                  <MetricsPanel metrics={metrics} nodes={nodes} />
                </div>
                <LeaderSchedulePanel />
              </div>
            </div>
          </div>
        </section>
        <div className="node-pulse-sidebar">
          <ChunkQueueDetailPanel />
        </div>
      </div>
      <TransactionFeedPanel />
      <PingLatencyPanel />
      <SystemLogPanel />
      <PidTelemetryPanel />
    </div>
  );
};

async function loadHistoricalLogs(
  range: { from: number; to: number },
  controller: AbortController,
): Promise<HistoricalEvent[]> {
  const baseUrl =
    process.env.NEXT_PUBLIC_API_URL ??
    process.env.BACKEND_URL ??
    "http://51.195.24.236:3000";
  const fromIso = formatLocalIso(range.from);
  const toIso = formatLocalIso(range.to);

  const [
    chunkResponse,
    routerResponse,
    pingResponse,
    bpfResponse,
    offcpuResponse,
    schedulerResponse,
    perfResponse,
    turboResponse,
    bftResponse,
    execResponse,
  ] = await Promise.all([
    fetch(
      `${baseUrl}/api/logs/chunk?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
    fetch(
      `${baseUrl}/api/logs/router?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
    fetch(
      `${baseUrl}/api/logs/ping?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
    fetch(
      `${baseUrl}/api/logs/bpf?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
    fetch(
      `${baseUrl}/api/logs/offcpu?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
    fetch(
      `${baseUrl}/api/logs/scheduler?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
    fetch(
      `${baseUrl}/api/logs/perf?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
    fetch(
      `${baseUrl}/api/logs/turbo?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
    fetch(
      `${baseUrl}/api/logs/bft?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
    fetch(
      `${baseUrl}/api/logs/exec?from=${encodeURIComponent(
        fromIso,
      )}&to=${encodeURIComponent(toIso)}`,
      { signal: controller.signal, cache: "no-store" },
    ),
  ]);

  if (!chunkResponse.ok) {
    throw new Error(`Failed to load chunk logs (${chunkResponse.status})`);
  }
  if (!routerResponse.ok) {
    throw new Error(`Failed to load router logs (${routerResponse.status})`);
  }
  if (!pingResponse.ok) {
    throw new Error(`Failed to load ping logs (${pingResponse.status})`);
  }
  if (!bpfResponse.ok) {
    throw new Error(`Failed to load bpf logs (${bpfResponse.status})`);
  }
  if (!offcpuResponse.ok) {
    throw new Error(`Failed to load offcpu logs (${offcpuResponse.status})`);
  }
  if (!schedulerResponse.ok) {
    throw new Error(
      `Failed to load scheduler logs (${schedulerResponse.status})`,
    );
  }
  if (!perfResponse.ok) {
    throw new Error(`Failed to load perf logs (${perfResponse.status})`);
  }
  if (!turboResponse.ok) {
    throw new Error(`Failed to load turbo logs (${turboResponse.status})`);
  }
  if (!bftResponse.ok) {
    throw new Error(`Failed to load bft logs (${bftResponse.status})`);
  }
  if (!execResponse.ok) {
    throw new Error(`Failed to load exec logs (${execResponse.status})`);
  }

  const chunkLogs = (await chunkResponse.json()) as unknown;
  const routerLogs = (await routerResponse.json()) as unknown;
  const pingLogs = (await pingResponse.json()) as unknown;
  const bpfLogs = (await bpfResponse.json()) as unknown;
  const offcpuLogs = (await offcpuResponse.json()) as unknown;
  const schedulerLogs = (await schedulerResponse.json()) as unknown;
  const perfLogs = (await perfResponse.json()) as unknown;
  const turboLogs = (await turboResponse.json()) as unknown;
  const bftLogs = (await bftResponse.json()) as unknown;
  const execLogs = (await execResponse.json()) as unknown;

  if (
    !Array.isArray(chunkLogs) ||
    !Array.isArray(routerLogs) ||
    !Array.isArray(pingLogs) ||
    !Array.isArray(bpfLogs) ||
    !Array.isArray(offcpuLogs) ||
    !Array.isArray(schedulerLogs) ||
    !Array.isArray(perfLogs) ||
    !Array.isArray(turboLogs) ||
    !Array.isArray(bftLogs) ||
    !Array.isArray(execLogs)
  ) {
    throw new Error("Unexpected historical payload shape");
  }

  const sortedChunks = [...chunkLogs]
    .map((entry) => ({
      timestamp: parseTimestamp(entry.timestamp),
      payload: entry as MonadChunkEvent,
    }))
    .sort((a, b) => a.timestamp - b.timestamp);

  const sortedRouter = [...routerLogs]
    .map((entry) => ({
      timestamp: parseTimestamp(entry.timestamp),
      payload: entry as OutboundRouterEvent,
    }))
    .sort((a, b) => a.timestamp - b.timestamp);

  const sortedPings = [...pingLogs]
    .map((entry) => ({
      timestamp: parseTimestamp((entry as any).timestamp),
      payload: entry as PingLatencyEvent,
    }))
    .sort((a, b) => a.timestamp - b.timestamp);

  const sortedBpf = [...bpfLogs]
    .map((entry) => ({
      timestamp: parseTimestamp((entry as any).timestamp),
      payload: entry as BpfTraceEvent,
    }))
    .sort((a, b) => a.timestamp - b.timestamp);

  const mergedSystemLogs = [...bftLogs, ...execLogs];
  const sortedSystem = [...mergedSystemLogs]
    .map((entry) => ({
      timestamp: parseTimestamp((entry as any).timestamp),
      payload: entry as SystemLogEvent,
    }))
    .sort((a, b) => a.timestamp - b.timestamp);

  const sortedOffcpu = [...offcpuLogs]
    .map((entry) => ({
      timestamp: parseTimestamp((entry as any).timestamp),
      payload: entry as OffCpuEvent,
    }))
    .sort((a, b) => a.timestamp - b.timestamp);

  const sortedScheduler = [...schedulerLogs]
    .map((entry) => ({
      timestamp: parseTimestamp((entry as any).timestamp),
      payload: entry as SchedulerEvent,
    }))
    .sort((a, b) => a.timestamp - b.timestamp);

  const sortedPerf = [...perfLogs]
    .map((entry) => ({
      timestamp: parseTimestamp((entry as any).timestamp),
      payload: entry as PerfStatEvent,
    }))
    .sort((a, b) => a.timestamp - b.timestamp);

  const sortedTurbo = [...turboLogs]
    .map((entry) => ({
      timestamp: parseTimestamp((entry as any).timestamp),
      payload: entry as TurboStatEvent,
    }))
    .sort((a, b) => a.timestamp - b.timestamp);

  const timeline = [
    ...sortedChunks.map((item) => ({
      type: "chunk" as const,
      timestamp: item.timestamp,
      payload: item.payload,
    })),
    ...sortedRouter.map((item) => ({
      type: "router" as const,
      timestamp: item.timestamp,
      payload: item.payload,
    })),
    ...sortedPings.map((item) => ({
      type: "ping" as const,
      timestamp: item.timestamp,
      payload: item.payload,
    })),
    ...sortedBpf.map((item) => ({
      type: "bpf" as const,
      timestamp: item.timestamp,
      payload: item.payload,
    })),
    ...sortedSystem.map((item) => ({
      type: "system" as const,
      timestamp: item.timestamp,
      payload: item.payload,
    })),
    ...sortedOffcpu.map((item) => ({
      type: "offcpu" as const,
      timestamp: item.timestamp,
      payload: item.payload,
    })),
    ...sortedScheduler.map((item) => ({
      type: "scheduler" as const,
      timestamp: item.timestamp,
      payload: item.payload,
    })),
    ...sortedPerf.map((item) => ({
      type: "perf" as const,
      timestamp: item.timestamp,
      payload: item.payload,
    })),
    ...sortedTurbo.map((item) => ({
      type: "turbo" as const,
      timestamp: item.timestamp,
      payload: item.payload,
    })),
  ].sort((a, b) => a.timestamp - b.timestamp);

  return timeline;
}

function parseTimestamp(value: unknown) {
  if (typeof value === "number") {
    return value;
  }
  if (typeof value === "string") {
    const parsed = Date.parse(value);
    if (!Number.isNaN(parsed)) {
      return parsed;
    }
    const asInt = Number.parseInt(value, 10);
    return Number.isNaN(asInt) ? 0 : asInt;
  }
  if (value instanceof Date) {
    return value.getTime();
  }
  return 0;
}

function ingestRouterEvent(event: OutboundRouterEvent) {
  const state = useNodePulseStore.getState();
  state.pushRouterEvent(event);
  const hash = getRouterEventHash(event);
  if (hash) {
    state.triggerChunkAssembly(hash);
  }
}

function getRouterEventHash(event: OutboundRouterEvent) {
  if (event.appMessageHash) {
    return event.appMessageHash;
  }
  if (
    event.data &&
    typeof event.data === "object" &&
    "appMessageHash" in event.data &&
    typeof (event.data as { appMessageHash?: string }).appMessageHash === "string"
  ) {
    return (event.data as { appMessageHash?: string }).appMessageHash;
  }
  return undefined;
}

function formatLocalIso(timestamp: number) {
  // Use UTC-based ISO string (without the trailing Z) so that
  // the query range matches the UTC timestamps stored in logs.
  return new Date(timestamp).toISOString().replace("Z", "");
}

export default MonadNodePulse;
