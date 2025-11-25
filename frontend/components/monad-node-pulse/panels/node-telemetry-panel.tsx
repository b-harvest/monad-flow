"use client";

import { useMemo } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";

export function NodeTelemetryPanel() {
  const selectedNodeId = useNodePulseStore((state) => state.selectedNodeId);
  const nodes = useNodePulseStore((state) => state.nodes);
  const telemetry = useNodePulseStore((state) =>
    selectedNodeId ? state.nodeTelemetry[selectedNodeId] : undefined,
  );
  const setSelectedNode = useNodePulseStore((state) => state.setSelectedNode);

  const node = useMemo(
    () => nodes.find((item) => item.id === selectedNodeId),
    [nodes, selectedNodeId],
  );

  if (!node || !node.isLocal) {
    return null;
  }

  return (
    <aside className="node-telemetry-panel glass-card">
      <header className="panel-header">
        <div>
          <span className="text-label">Node Deep Dive</span>
          <p className="text-title">{node.name}</p>
        </div>
        <button
          type="button"
          className="status-toggle ghost"
          onClick={() => setSelectedNode(null)}
        >
          Close
        </button>
      </header>
      <div className="telemetry-grid">
        <TelemetrySection title="BPF_TRACE" description="Function hooks">
          {telemetry?.bpfTrace.map((trace, index) => (
            <div key={`${trace.funcName}-${index}`} className="telemetry-card">
              <div className="telemetry-row">
                <span className="badge" data-variant="info">
                  {trace.eventType}
                </span>
                <span className="telemetry-timestamp">{trace.timestamp}</span>
              </div>
              <p className="telemetry-primary">{trace.funcName}</p>
              <p className="telemetry-detail">{trace.detail}</p>
              {trace.durationNs ? (
                <p className="telemetry-meta">{trace.durationNs} ns</p>
              ) : null}
            </div>
          ))}
        </TelemetrySection>

        <TelemetrySection title="SYSTEM_LOG" description="systemd journal">
          {telemetry?.systemLogs.map((log) => (
            <div key={`${log.service}-${log.timestamp}`} className="telemetry-card">
              <p className="telemetry-primary">{log.service}</p>
              <p className="telemetry-detail">{log.message}</p>
              <p className="telemetry-meta">{log.timestamp}</p>
            </div>
          ))}
        </TelemetrySection>

        <TelemetrySection title="OFF_CPU" description="blocked threads">
          {telemetry ? (
            <div className="telemetry-card">
              <p className="telemetry-primary">{telemetry.offCpu.processName}</p>
              <p className="telemetry-meta">{telemetry.offCpu.durationUs} µs off-CPU</p>
              <p className="telemetry-detail">
                {telemetry.offCpu.stack.join(" → ")}
              </p>
            </div>
          ) : (
            placeholderCard("Awaiting samples")
          )}
        </TelemetrySection>

        <TelemetrySection title="SCHEDULER" description="thread dynamics">
          {telemetry ? (
            <div className="telemetry-card">
              <p className="telemetry-primary">{telemetry.scheduler.threadName}</p>
              <p className="telemetry-detail">
                wait {telemetry.scheduler.waitDeltaMs} ms • run {telemetry.scheduler.runDeltaMs} ms
              </p>
              <p className="telemetry-meta">
                ctx switches {telemetry.scheduler.ctxSwitches}
              </p>
            </div>
          ) : (
            placeholderCard("Awaiting samples")
          )}
        </TelemetrySection>

        <TelemetrySection title="PERF_STAT" description="perf counters">
          {(telemetry?.perfStat ?? []).map((metric) => (
            <div key={metric.event} className="telemetry-card">
              <p className="telemetry-primary">{metric.event}</p>
              <p className="telemetry-detail">{metric.value}</p>
              <p className="telemetry-meta">
                {metric.metricVal ?? "—"} · {metric.runPct ?? "100%"}
              </p>
            </div>
          ))}
        </TelemetrySection>

        <TelemetrySection title="TURBO_STAT" description="power & freq">
          {telemetry ? (
            <div className="telemetry-card">
              <p className="telemetry-primary">Core {telemetry.turboStat.core}</p>
              <p className="telemetry-detail">
                {telemetry.turboStat.avgMhz} MHz · {telemetry.turboStat.busyPct}% busy
              </p>
              <p className="telemetry-meta">
                IPC {telemetry.turboStat.ipc} · {telemetry.turboStat.pkgWatt} W pkg
              </p>
            </div>
          ) : (
            placeholderCard("Awaiting samples")
          )}
        </TelemetrySection>
      </div>
    </aside>
  );
}

function TelemetrySection({
  title,
  description,
  children,
}: {
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <section className="telemetry-section">
      <div className="telemetry-section-head">
        <span className="text-label">{title}</span>
        <span className="telemetry-section-desc">{description}</span>
      </div>
      <div className="telemetry-section-body">{children}</div>
    </section>
  );
}

function placeholderCard(message: string) {
  return <div className="telemetry-card telemetry-placeholder">{message}</div>;
}
