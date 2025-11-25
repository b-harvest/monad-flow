"use client";

import { Suspense, useState } from "react";
import { NodePulseScene } from "./node-pulse-scene";
import { useMockPulseStream } from "@/lib/hooks/use-mock-pulse-stream";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { StatusBar } from "./status/status-bar";
import { MetricsPanel } from "./panels/metrics-panel";
import { PlaybackControls } from "./panels/playback-controls";
import { AlertToast } from "./status/alert-toast";
import { NodeTelemetryPanel } from "./panels/node-telemetry-panel";
import { EventLogPanel } from "./panels/event-log-panel";
import { CommandNav } from "./top-nav";
import { OutboundRouterDebugger } from "../debug/outbound-router-debugger";
import { SocketLivePanel } from "../debug/socket-live-panel";
import "@/lib/storage/bpf-trace-cache";
import { useBpfTraceStream } from "@/lib/socket/use-bpf-trace-stream";
import { useSystemLogStream } from "@/lib/socket/use-system-log-stream";
import { usePerfStatStream } from "@/lib/socket/use-perf-stat-stream";
import { useOffCpuStream } from "@/lib/socket/use-off-cpu-stream";
import { useSchedulerStream } from "@/lib/socket/use-scheduler-stream";
import { useTurboStatStream } from "@/lib/socket/use-turbo-stat-stream";
import { useMonadChunkStream } from "@/lib/socket/use-monad-chunk-stream";

const MonadNodePulse = () => {
  useMockPulseStream();
  useBpfTraceStream();
  useSystemLogStream();
  useOffCpuStream();
  usePerfStatStream();
  useSchedulerStream();
  useTurboStatStream();
  useMonadChunkStream();
  const [socketPanelOpen, setSocketPanelOpen] = useState(false);
  const [routerDebugOpen, setRouterDebugOpen] = useState(false);
  const [socketDebugOpen, setSocketDebugOpen] = useState(false);
  const metrics = useNodePulseStore((state) => state.metrics);
  const nodes = useNodePulseStore((state) => state.nodes);
  const eventLog = useNodePulseStore((state) => state.eventLog);
  const playback = useNodePulseStore((state) => state.playback);
  const setPlayback = useNodePulseStore((state) => state.setPlayback);
  const lastEventTimestamp = useNodePulseStore(
    (state) => state.lastEventTimestamp,
  );
  const alert = useNodePulseStore((state) => state.alert);
  const setAlert = useNodePulseStore((state) => state.setAlert);

  return (
    <div className="node-pulse-shell">
      <CommandNav
        streamOpen={socketPanelOpen}
        onToggleStream={() => setSocketPanelOpen((prev) => !prev)}
        debugOpen={routerDebugOpen}
        onToggleDebug={() => setRouterDebugOpen((prev) => !prev)}
        socketOpen={socketDebugOpen}
        onToggleSocket={() => setSocketDebugOpen((prev) => !prev)}
      />
      <div className="node-pulse-layout">
        <section className="node-pulse-stage">
          <Suspense fallback={<div className="visualization-canvas" />}>
            <NodePulseScene />
          </Suspense>
          <StatusBar
            connectionStatus={metrics.connectionStatus}
            lastEventTimestamp={lastEventTimestamp}
            playbackMode={playback.mode}
            liveAvailable={playback.liveAvailable}
          />
          <NodeTelemetryPanel />
          {alert ? (
            <AlertToast alert={alert} onDismiss={() => setAlert(null)} />
          ) : null}
        </section>
      </div>
      <div className="node-pulse-foot">
        <MetricsPanel metrics={metrics} nodes={nodes} />
        <PlaybackControls playback={playback} onChange={setPlayback} />
      </div>
      {socketPanelOpen ? (
        <div className="socket-panel">
          <EventLogPanel events={eventLog} />
        </div>
      ) : null}
      <OutboundRouterDebugger
        open={routerDebugOpen}
        onClose={() => setRouterDebugOpen(false)}
      />
      <SocketLivePanel open={socketDebugOpen} onClose={() => setSocketDebugOpen(false)} />
    </div>
  );
};

export default MonadNodePulse;
