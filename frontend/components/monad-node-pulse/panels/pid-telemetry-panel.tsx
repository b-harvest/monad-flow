"use client";

import { useEffect, useMemo, useState, useSyncExternalStore } from "react";
import {
  getSchedulerEvents,
  subscribeToSchedulerEvents,
} from "@/lib/storage/scheduler-cache";
import type { SchedulerEvent } from "@/lib/api/scheduler";
import {
  getPerfStatEvents,
  subscribeToPerfStatEvents,
} from "@/lib/storage/perf-stat-cache";
import type { PerfStatEvent } from "@/lib/api/perf-stat";
import {
  getOffCpuEvents,
  subscribeToOffCpuEvents,
} from "@/lib/storage/off-cpu-cache";
import {
  getTurboStatEvents,
  subscribeToTurboStatEvents,
} from "@/lib/storage/turbo-stat-cache";
import type { TurboStatEvent } from "@/lib/api/turbo-stat";

interface PidSnapshot {
  pid: string;
  scheduler?: {
    threadName: string;
    waitMs: number;
    runMs: number;
    ctxSwitches: number;
  };
  perfMetrics?: Array<{
    event: string;
    value: string;
    meta?: string;
  }>;
  offCpu?: {
    processName: string;
    durationUs: number;
    timestamp: number;
  };
  updatedAt: number;
}

const sortSnapshots = (rows: PidSnapshot[]) =>
  rows.sort((a, b) => {
    const aNum = Number.parseInt(a.pid, 10);
    const bNum = Number.parseInt(b.pid, 10);
    if (Number.isNaN(aNum) && Number.isNaN(bNum)) {
      return a.pid.localeCompare(b.pid);
    }
    if (Number.isNaN(aNum)) return 1;
    if (Number.isNaN(bNum)) return -1;
    return aNum - bNum;
  });

export function PidTelemetryPanel() {
  const [hydrated, setHydrated] = useState(false);
  const schedulerEvents = useSyncExternalStore(
    subscribeToSchedulerEvents,
    getSchedulerEvents,
    getSchedulerEvents,
  );
  const perfEvents = useSyncExternalStore(
    subscribeToPerfStatEvents,
    getPerfStatEvents,
    getPerfStatEvents,
  );
  const offCpuEvents = useSyncExternalStore(
    subscribeToOffCpuEvents,
    getOffCpuEvents,
    getOffCpuEvents,
  );
  const turboEvents = useSyncExternalStore(
    subscribeToTurboStatEvents,
    getTurboStatEvents,
    getTurboStatEvents,
  );

  useEffect(() => {
    setHydrated(true);
  }, []);

  const snapshots = useMemo(() => {
    const map = new Map<string, PidSnapshot>();

    const ensurePid = (pid?: string | number) => {
      if (!pid && pid !== 0) {
        return null;
      }
      const key = String(pid);
      if (!map.has(key)) {
        map.set(key, { pid: key, updatedAt: 0 });
      }
      return map.get(key)!;
    };

    schedulerEvents.slice(-50).forEach((event) => {
      const pid =
        (event as SchedulerEvent & { pid?: string | number }).pid ??
        event.main_pid;
      const entry = ensurePid(pid);
      if (!entry) return;
      const ctxSwitches =
        typeof event.ctx_switches === "number"
          ? event.ctx_switches
          : Number.parseInt(event.ctx_switches, 10) || 0;
      const timestamp = Date.parse(event.timestamp) || 0;
      entry.scheduler = {
        threadName: event.thread_name,
        waitMs: event.wait_delta_ms,
        runMs: event.run_delta_ms,
        ctxSwitches,
      };
      entry.updatedAt = Math.max(entry.updatedAt, timestamp);
    });

    perfEvents.slice(-30).forEach((event) => {
      const pid = (event as PerfStatEvent & { pid?: string | number }).pid;
      const entry = ensurePid(pid);
      if (!entry) return;
      const timestamp = Date.parse(event.timestamp) || 0;
      entry.perfMetrics = event.metrics.slice(0, 3).map((metric) => ({
        event: metric.event,
        value: metric.unit ? `${metric.value} ${metric.unit}` : metric.value,
        meta: metric.metric_val ?? metric.run_pct ?? undefined,
      }));
      entry.updatedAt = Math.max(entry.updatedAt, timestamp);
    });

    offCpuEvents.slice(-12).forEach((event) => {
      const entry = ensurePid(event.pid);
      if (!entry) return;
      const timestamp = Date.parse(event.timestamp) || 0;
      entry.offCpu = {
        processName: event.process_name,
        durationUs: event.duration_us,
        timestamp,
      };
      entry.updatedAt = Math.max(entry.updatedAt, timestamp);
    });

    return sortSnapshots(Array.from(map.values()));
  }, [schedulerEvents, perfEvents, offCpuEvents]);

  return (
    <section className="pid-telemetry-panel hud-panel">
      <header className="panel-header">
        <div>
          <span className="text-label">Process Telemetry</span>
        </div>
      </header>
      <TurboStatStrip events={hydrated ? turboEvents.slice(-4) : []} />
      <div className="pid-grid-wrapper">
        {!hydrated || snapshots.length === 0 ? (
          <p className="pid-placeholder">Waiting for process samples…</p>
        ) : (
          <div className="pid-grid">
            {snapshots.map((snapshot) => (
              <div key={snapshot.pid} className="pid-card">
                <div className="pid-card-head">
                  <div>
                    <span className="text-label">PID</span>
                    <span className="text-number">{snapshot.pid}</span>
                  </div>
                  <span className="pid-updated">
                    {snapshot.updatedAt ? formatRelative(snapshot.updatedAt) : "—"}
                  </span>
                </div>
                <div className="pid-card-section">
                  <span className="text-label">Scheduler</span>
                  {snapshot.scheduler ? (
                    <div className="pid-values">
                      <PidValue
                        label="Wait"
                        value={`${snapshot.scheduler.waitMs.toFixed(2)} ms`}
                      />
                      <PidValue
                        label="Run"
                        value={`${snapshot.scheduler.runMs.toFixed(2)} ms`}
                      />
                      <PidValue
                        label="Ctx Switches"
                        value={snapshot.scheduler.ctxSwitches.toLocaleString()}
                        hint={snapshot.scheduler.threadName}
                      />
                    </div>
                  ) : (
                    <p className="pid-placeholder">waiting…</p>
                  )}
                </div>
                <div className="pid-card-section">
                  <span className="text-label">Perf</span>
                  {snapshot.perfMetrics ? (
                    <div className="pid-values">
                      {snapshot.perfMetrics.map((metric) => (
                        <PidValue
                          key={`${snapshot.pid}-${metric.event}`}
                          label={metric.event}
                          value={metric.value}
                          hint={metric.meta}
                        />
                      ))}
                    </div>
                  ) : (
                    <p className="pid-placeholder">waiting…</p>
                  )}
                </div>
                <div className="pid-card-section">
                  <span className="text-label">Off-CPU</span>
                  {snapshot.offCpu ? (
                    <div className="pid-values">
                      <PidValue
                        label={snapshot.offCpu.processName}
                        value={`${snapshot.offCpu.durationUs.toLocaleString()} µs`}
                        hint={formatRelative(snapshot.offCpu.timestamp)}
                      />
                    </div>
                  ) : (
                    <p className="pid-placeholder">waiting…</p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}

function formatRelative(timestamp: number) {
  if (!timestamp) return "—";
  const diff = Date.now() - timestamp;
  const seconds = Math.round(diff / 1000);
  if (seconds < 1) return "just now";
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.round(minutes / 60);
  return `${hours}h ago`;
}

function PidValue({
  label,
  value,
  hint,
}: {
  label: string;
  value: string;
  hint?: string;
}) {
  return (
    <div className="pid-value">
      <span className="pid-value-label">{label}</span>
      <strong className="pid-value-number">{value}</strong>
      {hint ? <span className="pid-bar-hint">{hint}</span> : null}
    </div>
  );
}

function TurboStatStrip({ events }: { events: TurboStatEvent[] }) {
  if (events.length === 0) {
    return (
      <div className="pid-panel-top">
        <p className="pid-placeholder">Waiting for turbo samples…</p>
      </div>
    );
  }
  const latest = events[events.length - 1];
  return (
    <div className="pid-panel-top">
      <div className="pid-turbo-grid">
        <div className="pid-card turbo-card">
          <span className="text-label">Turbo Stat</span>
          <div className="pid-values">
            <PidValue label="Core" value={latest.core} />
            <PidValue label="Avg MHz" value={latest.avg_mhz.toFixed(1)} />
            <PidValue label="Busy %" value={`${latest.busy_pct.toFixed(1)}%`} />
            <PidValue label="IPC" value={latest.ipc.toFixed(2)} />
            <PidValue label="Pkg Watt" value={`${latest.pkg_watt.toFixed(2)} W`} />
            <PidValue label="Core Watt" value={`${latest.cor_watt.toFixed(2)} W`} />
          </div>
        </div>
      </div>
    </div>
  );
}
