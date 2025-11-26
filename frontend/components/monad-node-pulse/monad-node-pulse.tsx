"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { NodePulseMap } from "./node-pulse-map";
import { useMockPulseStream } from "@/lib/hooks/use-mock-pulse-stream";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { MetricsPanel } from "./panels/metrics-panel";
import { ChunkQueueDetailPanel } from "./panels/chunk-queue-detail-panel";
import { AlertToast } from "./status/alert-toast";
import { NodeTelemetryPanel } from "./panels/node-telemetry-panel";
import { PidTelemetryPanel } from "./panels/pid-telemetry-panel";
import { SystemLogPanel } from "./panels/system-log-panel";
import { CommandNav } from "./top-nav";
import "@/lib/storage/bpf-trace-cache";
import { useBpfTraceStream } from "@/lib/socket/use-bpf-trace-stream";
import { useSystemLogStream } from "@/lib/socket/use-system-log-stream";
import { usePerfStatStream } from "@/lib/socket/use-perf-stat-stream";
import { useOffCpuStream } from "@/lib/socket/use-off-cpu-stream";
import { useSchedulerStream } from "@/lib/socket/use-scheduler-stream";
import { useTurboStatStream } from "@/lib/socket/use-turbo-stat-stream";
import { useMonadChunkStream } from "@/lib/socket/use-monad-chunk-stream";
import { useOutboundRouterStream } from "@/lib/socket/use-outbound-router-stream";

const MonadNodePulse = () => {
  useMockPulseStream();
  useBpfTraceStream();
  useSystemLogStream();
  useOffCpuStream();
  usePerfStatStream();
  useSchedulerStream();
  useTurboStatStream();
  useMonadChunkStream();
  useOutboundRouterStream();
  const metrics = useNodePulseStore((state) => state.metrics);
  const nodes = useNodePulseStore((state) => state.nodes);
  const playback = useNodePulseStore((state) => state.playback);
  const setPlayback = useNodePulseStore((state) => state.setPlayback);
  const alert = useNodePulseStore((state) => state.alert);
  const setAlert = useNodePulseStore((state) => state.setAlert);

  const [processTelemetryVisible, setProcessTelemetryVisible] = useState(true);
  const [mapShellEl, setMapShellEl] = useState<HTMLDivElement | null>(null);
  const mapShellRef = useCallback((node: HTMLDivElement | null) => {
    setMapShellEl(node);
  }, []);
  const [mapShellBounds, setMapShellBounds] = useState<DOMRect | null>(null);
  const [hasMounted, setHasMounted] = useState(false);

  useEffect(() => {
    setHasMounted(true);
  }, []);

  useEffect(() => {
    if (!mapShellEl) return;
    const measure = () => {
      setMapShellBounds(mapShellEl.getBoundingClientRect());
    };
    measure();
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
  }, [mapShellEl]);

  return (
    <div className="node-pulse-shell">
      <CommandNav
        connectionStatus={metrics.connectionStatus}
        playback={playback}
        onPlaybackChange={setPlayback}
        processTelemetryVisible={processTelemetryVisible}
        onToggleProcessTelemetry={() =>
          setProcessTelemetryVisible((prev) => !prev)
        }
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
              <MetricsPanel metrics={metrics} nodes={nodes} />
            </div>
          </div>
          {alert ? (
            <AlertToast alert={alert} onDismiss={() => setAlert(null)} />
          ) : null}
        </section>
        <div className="node-pulse-sidebar">
          <ChunkQueueDetailPanel />
        </div>
      </div>
      <SystemLogPanel />
      {processTelemetryVisible && hasMounted && mapShellBounds
        ? createPortal(
            <div
              className="process-telemetry-overlay"
              style={{
                left: mapShellBounds.left,
                top: mapShellBounds.top,
                width: mapShellBounds.width,
                height: mapShellBounds.height,
              }}
            >
              <PidTelemetryPanel />
            </div>,
            document.body,
          )
        : null}
    </div>
  );
};

export default MonadNodePulse;
