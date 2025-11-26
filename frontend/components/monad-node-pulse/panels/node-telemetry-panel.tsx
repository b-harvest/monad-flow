"use client";

import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  useSyncExternalStore,
} from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import {
  getBpfTraceEvents,
  subscribeToBpfTraceEvents,
} from "@/lib/storage/bpf-trace-cache";

const UPDATED_TIME_FORMATTER = new Intl.DateTimeFormat("en-GB", {
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hour12: false,
  timeZone: "UTC",
});

export function NodeTelemetryPanel() {
  const nodes = useNodePulseStore((state) => state.nodes);
  const [hydrated, setHydrated] = useState(false);
  const bpfTraceEvents = useSyncExternalStore(
    subscribeToBpfTraceEvents,
    getBpfTraceEvents,
    getBpfTraceEvents,
  );

  useEffect(() => {
    setHydrated(true);
  }, []);

  const node = useMemo(
    () => nodes.find((item) => item.isLocal) ?? nodes[0],
    [nodes],
  );

  const bpfSeries = useMemo(() => {
    const map = new Map<
      string,
      { funcName: string; values: { timestamp: number; value: number }[] }
    >();
    const recent = bpfTraceEvents.slice(-80);
    recent.forEach((event) => {
      const duration = Number(event.duration_ns ?? "0");
      if (!Number.isFinite(duration) || duration <= 0) {
        return;
      }
      const timestamp = Date.parse(event.timestamp) || Date.now();
      const next = map.get(event.func_name) ?? {
        funcName: event.func_name,
        values: [],
      };
      next.values = [
        ...next.values,
        { timestamp, value: duration },
      ].slice(-20);
      map.set(event.func_name, next);
    });
    return Array.from(map.values())
      .sort((a, b) => a.funcName.localeCompare(b.funcName))
      .slice(0, 6);
  }, [bpfTraceEvents]);

  const lastUpdated = useMemo(() => {
    const ts = bpfTraceEvents[bpfTraceEvents.length - 1]?.timestamp;
    if (!ts) return null;
    const date = new Date(ts);
    return Number.isNaN(date.getTime()) ? null : date;
  }, [bpfTraceEvents]);

  return (
    <aside className="node-telemetry-panel">
      <div className="telemetry-content">
        <header className="panel-header telemetry-panel-head">
          <div>
            <span className="text-label">Node Deep Dive</span>
            <p className="text-title">{node?.name ?? "Local Validator"}</p>
          </div>
          <div className="telemetry-updated">
            <span className="text-label">Updated</span>
            <span className="text-number">
              {hydrated && lastUpdated
                ? UPDATED_TIME_FORMATTER.format(lastUpdated)
                : "waiting…"}
            </span>
          </div>
        </header>

        <div className="telemetry-grid">
          <TelemetrySection title="BPF_TRACE" description="Function hooks">
            {hydrated && bpfSeries.length > 0 ? (
              <div className="bpf-spark-grid">
                {bpfSeries.map((series) => (
                  <div key={series.funcName} className="telemetry-card bpf-spark-card">
                    <p className="telemetry-primary bpf-card-title">{series.funcName}</p>
                    <div className="bpf-chart">
                      <TimelineSparkline
                        label="duration"
                        values={series.values}
                      />
                      <div className="bpf-duration-meta">
                        <span className="text-label">Duration</span>
                        <span className="text-number">
                          {series.values.length > 0
                            ? `${series.values[series.values.length - 1].value.toLocaleString()} ns`
                            : "—"}
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="sparkline sparkline-placeholder">
                waiting for BPF events…
              </div>
            )}
          </TelemetrySection>
        </div>
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
        <div>
          <span className="text-label">{title}</span>
          <span className="telemetry-section-desc">{description}</span>
        </div>
      </div>
      <div className="telemetry-section-body">{children}</div>
    </section>
  );
}

function TimelineSparkline({
  values,
  label,
}: {
  values: { timestamp: number; value: number }[];
  label: string;
}) {
  const [containerRef, width] = useSparklineWidth();
  const height = 320;
  if (values.length === 0) {
    return (
      <div className="sparkline sparkline-placeholder" ref={containerRef}>
        waiting for {label} samples…
      </div>
    );
  }
  const min = 0;
  const max = 500_000_000;
  const range = max - min;
  const minTs = values[0]?.timestamp ?? 0;
  const maxTs = values[values.length - 1]?.timestamp ?? minTs + 1;
  const tsRange = maxTs - minTs || 1;
  const axisPadding = 12;
  const originX = axisPadding;
  const originY = height - axisPadding;
  const chartWidth = width - axisPadding * 2;
  const chartHeight = originY - axisPadding;
  const coords = values.map((point) => {
    const x = originX + ((point.timestamp - minTs) / tsRange) * chartWidth;
    const y =
      originY - ((point.value - min) / range) * chartHeight;
    return { x, y };
  });
  return (
    <div className="sparkline" ref={containerRef}>
      <svg
        viewBox={`0 0 ${width} ${height}`}
        role="img"
        aria-label={`${label} sparkline`}
      >
        <line
          x1={originX}
          y1={axisPadding}
          x2={originX}
          y2={originY}
          stroke="rgba(133, 230, 255, 0.25)"
          strokeWidth="1"
        />
        <line
          x1={axisPadding}
          y1={originY}
          x2={width - axisPadding}
          y2={originY}
          stroke="rgba(133, 230, 255, 0.25)"
          strokeWidth="1"
        />
        {coords.length > 1 ? (
          <path
            d={coords
              .map((point, index) =>
                `${index === 0 ? "M" : "L"}${point.x.toFixed(1)},${point.y.toFixed(1)}`,
              )
              .join("")}
            fill="none"
            stroke="rgba(133, 230, 255, 0.65)"
            strokeWidth="1.5"
            strokeLinecap="round"
          />
        ) : null}
        {coords.map((point, index) => (
          <circle
            key={`${point.x}-${point.y}-${index}`}
            cx={point.x}
            cy={point.y}
            r="3"
            fill="rgb(133, 230, 255)"
            stroke="rgba(8, 6, 20, 0.9)"
            strokeWidth="1"
          />
        ))}
      </svg>
    </div>
  );
}

function useSparklineWidth() {
  const [node, setNode] = useState<HTMLDivElement | null>(null);
  const [width, setWidth] = useState(320);
  const ref = useCallback((instance: HTMLDivElement | null) => {
    setNode(instance);
  }, []);
  useEffect(() => {
    if (!node) return;
    const measure = () => {
      setWidth(node.clientWidth || 320);
    };
    measure();
    if (typeof ResizeObserver !== "undefined") {
      const observer = new ResizeObserver(measure);
      observer.observe(node);
      return () => observer.disconnect();
    }
    window.addEventListener("resize", measure);
    return () => window.removeEventListener("resize", measure);
  }, [node]);
  return [ref, Math.max(width, 120)] as const;
}
