"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import io, { Socket } from "socket.io-client";
import type { SocketEventRecord } from "@/types/monad";
import { SOCKET_EVENT_NAMES, defaultSocketEndpoint } from "@/lib/socket/config";

interface SocketLivePanelProps {
  open: boolean;
  onClose: () => void;
}

const MAX_LOG = 80;

export function SocketLivePanel({ open, onClose }: SocketLivePanelProps) {
  const [endpoint, setEndpoint] = useState(defaultSocketEndpoint);
  const [socket, setSocket] = useState<Socket | null>(null);
  const [status, setStatus] = useState("Not connected");
  const [filter, setFilter] = useState<string>("ALL");
  const [log, setLog] = useState<SocketEventRecord[]>([]);
  const logRef = useRef<HTMLTextAreaElement>(null);

  const filteredLog = useMemo(() => {
    if (filter === "ALL") return log;
    return log.filter((entry) => entry.event === filter);
  }, [filter, log]);

  useEffect(() => {
    if (!open) return;
    if (logRef.current) {
      logRef.current.scrollTop = 0;
    }
  }, [filteredLog, open]);

  useEffect(() => {
    return () => {
      if (socket) {
        socket.disconnect();
      }
    };
  }, [socket]);

  const connect = () => {
    if (!endpoint.trim()) {
      setStatus("Enter a Socket.IO URL first.");
      return;
    }
    if (socket?.connected) {
      setStatus("Already connected.");
      return;
    }
    setStatus("Connecting…");
    const client = io(endpoint, {
      transports: ["polling", "websocket"],
    });
    client.on("connect", () => {
      setStatus(`Connected (${client.id})`);
      appendLog("INFO", `Connected to ${endpoint}`);
      SOCKET_EVENT_NAMES.forEach((event) => {
        client.on(event, (payload) => {
          appendLog(event, JSON.stringify(payload, null, 2));
        });
      });
      client.onAny((event, payload) => {
        if (!SOCKET_EVENT_NAMES.includes(event)) {
          appendLog(event, JSON.stringify(payload, null, 2));
        }
      });
    });
    client.on("connect_error", (error) => {
      setStatus(`Connect error: ${error.message}`);
    });
    client.on("disconnect", (reason) => {
      setStatus(`Disconnected: ${reason}`);
      appendLog("INFO", `Disconnected (${reason})`);
    });
    setSocket(client);
  };

  const disconnect = () => {
    if (!socket) return;
    socket.disconnect();
    setSocket(null);
    setStatus("Disconnected");
  };

  const appendLog = (event: string, payload: string) => {
    setLog((prev) => {
      const entry: SocketEventRecord = {
        event,
        payload,
        timestamp: Date.now(),
      };
      const next = [entry, ...prev];
      return next.slice(0, MAX_LOG);
    });
  };

  if (!open) {
    return null;
  }

  return (
    <div className="socket-live-panel glass-card">
      <header className="debug-header">
        <div>
          <span className="text-label">Socket.IO Live Inspector</span>
          <p className="text-title">{status}</p>
        </div>
        <div className="debug-actions">
          <button
            type="button"
            className="status-toggle"
            onClick={connect}
          >
            Connect
          </button>
          <button
            type="button"
            className="status-toggle"
            onClick={disconnect}
          >
            Disconnect
          </button>
          <button
            type="button"
            className="status-toggle ghost"
            onClick={onClose}
          >
            Close
          </button>
        </div>
      </header>
      <div className="socket-grid">
        <div className="socket-card">
          <label className="text-label" htmlFor="socket-endpoint">
            Endpoint
          </label>
          <input
            id="socket-endpoint"
            value={endpoint}
            onChange={(event) => setEndpoint(event.target.value)}
            placeholder="http://51.195.24.236:3000"
          />
          <p className="socket-tip">
            연결은 모니터링 전용이에요. 프론트는 서버 이벤트를 전송하지 않습니다.
          </p>
        </div>
        <div className="socket-card">
          <div className="socket-row">
            <label className="text-label" htmlFor="socket-filter">
              Filter
            </label>
            <select
              id="socket-filter"
              value={filter}
              onChange={(event) => setFilter(event.target.value)}
            >
              {["ALL", ...SOCKET_EVENT_NAMES].map((event) => (
                <option key={event} value={event}>
                  {event}
                </option>
              ))}
            </select>
          </div>
          <textarea
            ref={logRef}
            readOnly
            value={filteredLog
              .map(
                (entry) =>
                  `[${new Date(entry.timestamp).toLocaleTimeString()}] ${
                    entry.event
                  }\n${entry.payload}`,
              )
              .join("\n\n---\n\n")}
            placeholder="Incoming events will appear here..."
          />
        </div>
      </div>
    </div>
  );
}
