"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  fetchRouterLogs,
  type RouterLogEntry,
} from "@/lib/api/outbound-router";

interface OutboundRouterDebuggerProps {
  open: boolean;
  onClose: () => void;
}

const MAX_CAPTURE = 80;

export function OutboundRouterDebugger({
  open,
  onClose,
}: OutboundRouterDebuggerProps) {
  const [entries, setEntries] = useState<RouterLogEntry[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [issues, setIssues] = useState<string[]>([]);
  const [autoCapture, setAutoCapture] = useState(false);

  const selectedEntry = useMemo(
    () => entries.find((entry) => entry._id === selectedId) ?? entries[0],
    [entries, selectedId],
  );

  const captureSnapshot = useCallback(
    async (replace = false) => {
      setLoading(true);
      try {
        const { entries: data, issues: warnings } = await fetchRouterLogs({
          windowMs: 1000 * 60 * 2,
          limit: 50,
        });
        setEntries((prev) => {
          const merged = replace ? data : [...data, ...prev];
          return merged.slice(0, MAX_CAPTURE);
        });
        if (warnings.length > 0) {
          setIssues(warnings);
        } else {
          setIssues([]);
        }
        if (!selectedId && data.length > 0) {
          setSelectedId(data[0]._id);
        }
      } catch (error) {
        setIssues([(error as Error).message]);
      } finally {
        setLoading(false);
      }
    },
    [selectedId],
  );

  useEffect(() => {
    if (!open) return;
    if (entries.length === 0) {
      captureSnapshot(true);
    }
  }, [open, entries.length, captureSnapshot]);

  useEffect(() => {
    if (!autoCapture || !open) return;
    const interval = setInterval(() => {
      captureSnapshot(false);
    }, 4000);
    return () => clearInterval(interval);
  }, [autoCapture, open, captureSnapshot]);

  if (!open) {
    return null;
  }

  return (
    <div className="debug-drawer glass-card">
      <div className="debug-header">
        <div>
          <span className="text-label">DTO Inspector</span>
          <p className="text-title">Outbound Router Payloads</p>
        </div>
        <div className="debug-actions">
          <button
            type="button"
            className={`status-toggle ${autoCapture ? "active" : ""}`}
            onClick={() => setAutoCapture((prev) => !prev)}
          >
            {autoCapture ? "Pause Auto" : "Auto Sample"}
          </button>
          <button
            type="button"
            className="status-toggle"
            onClick={() => captureSnapshot(true)}
            disabled={loading}
          >
            {loading ? "Fetching…" : "Fetch Snapshot"}
          </button>
          <button
            type="button"
            className="status-toggle ghost"
            onClick={onClose}
          >
            Close
          </button>
        </div>
      </div>
      <div className="debug-body">
        <div className="debug-list">
          {entries.length === 0 ? (
            <p className="router-panel-placeholder">
              No router samples captured yet.
            </p>
          ) : (
            entries.map((entry) => (
              <button
                type="button"
                key={entry._id}
                className={`router-row ${selectedEntry?._id === entry._id ? "active" : ""}`}
                onClick={() => setSelectedId(entry._id)}
              >
                <p className="router-row-label">
                  {formatTimestamp(entry.timestamp)} · msgType{" "}
                  {entry.messageType ?? "?"}
                </p>
              </button>
            ))
          )}
        </div>
        <div className="debug-detail">
          {selectedEntry ? (
            <>
              <div className="router-detail-grid">
                <DetailItem label="Event ID" value={selectedEntry._id} />
                <DetailItem
                  label="Type"
                  value={selectedEntry.type ?? "Unknown"}
                />
                <DetailItem
                  label="Timestamp"
                  value={formatTimestamp(selectedEntry.timestamp)}
                />
                <DetailItem
                  label="Message Type"
                  value={String(selectedEntry.messageType ?? "—")}
                />
                <DetailItem
                  label="App Hash"
                  value={selectedEntry.appMessageHash ?? "—"}
                />
              </div>
              <div className="debug-json">
                <pre>{JSON.stringify(selectedEntry.data, null, 2)}</pre>
              </div>
            </>
          ) : (
            <p className="router-panel-placeholder">
              Select a sample to inspect its payload.
            </p>
          )}
        </div>
      </div>
      {issues.length > 0 ? (
        <div className="debug-issues">
          <span className="text-label">Validation Notes</span>
          <ul>
            {issues.map((issue) => (
              <li key={issue}>{issue}</li>
            ))}
          </ul>
        </div>
      ) : null}
    </div>
  );
}

function DetailItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="router-detail-item">
      <span className="text-label">{label}</span>
      <span className="text-number">{value}</span>
    </div>
  );
}

function formatTimestamp(value: number | string | undefined) {
  if (value === undefined) return "—";
  if (typeof value === "number") {
    return new Date(value).toLocaleString();
  }
  return new Date(value).toLocaleString();
}
