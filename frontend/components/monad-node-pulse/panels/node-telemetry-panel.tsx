"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";

const UPDATED_TIME_FORMATTER = new Intl.DateTimeFormat("en-GB", {
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hour12: false,
  timeZone: "UTC",
});

export function NodeTelemetryPanel() {
  const nodes = useNodePulseStore((state) => state.nodes);
  const bpfTraceEvents = useNodePulseStore((state) => state.bpfTraceEvents);
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    setHydrated(true);
  }, []);

  const node = useMemo(
    () => nodes.find((item) => item.isLocal) ?? nodes[0],
    [nodes],
  );

  const bpfSeries = useMemo(() => {
    return Object.values(bpfTraceEvents)
      .map((events) => {
        if (events.length === 0) return null;
        const funcName = events[0].func_name;
        const values = events.map((e) => {
          const timestamp = Date.parse(e.timestamp) || Date.now();
          const duration = Number(e.duration_ns ?? "0");
          return { timestamp, value: duration };
        });
        return { funcName, values };
      })
      .filter((item) => item !== null)
      .sort((a, b) => a!.funcName.localeCompare(b!.funcName));
  }, [bpfTraceEvents]);

  const lastUpdated = useMemo(() => {
    let maxTs = 0;
    Object.values(bpfTraceEvents).forEach((events) => {
      const last = events[events.length - 1];
      if (last) {
        const ts = Date.parse(last.timestamp);
        if (!Number.isNaN(ts) && ts > maxTs) {
          maxTs = ts;
        }
      }
    });
    return maxTs > 0 ? new Date(maxTs) : null;
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
