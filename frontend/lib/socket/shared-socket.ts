"use client";

import { io, Socket } from "socket.io-client";

type SharedState = {
  socket: Socket | null;
  endpoint: string | null;
  refs: number;
  idleTimer: number | null;
};

const globalKey = "__monad_flow_shared_socket__";
const globalState: SharedState =
  (globalThis as unknown as Record<string, SharedState>)[globalKey] ?? {
    socket: null,
    endpoint: null,
    refs: 0,
    idleTimer: null,
  };

(globalThis as unknown as Record<string, SharedState>)[globalKey] = globalState;

function scheduleIdleDisconnect() {
  if (typeof window === "undefined") {
    globalState.socket?.disconnect();
    globalState.socket = null;
    globalState.endpoint = null;
    globalState.idleTimer = null;
    return;
  }
  if (globalState.idleTimer) {
    window.clearTimeout(globalState.idleTimer);
    globalState.idleTimer = null;
  }
  if (!globalState.socket) {
    return;
  }
  globalState.idleTimer = window.setTimeout(() => {
    globalState.socket?.disconnect();
    globalState.socket = null;
    globalState.endpoint = null;
    globalState.idleTimer = null;
  }, 30_000);
}

export function acquireSocket(endpoint: string): Socket {
  if (!endpoint) {
    throw new Error("Socket endpoint is required");
  }
  if (globalState.socket) {
    if (globalState.endpoint !== endpoint) {
      globalState.socket.disconnect();
      globalState.socket = null;
      globalState.endpoint = null;
    } else {
      if (globalState.idleTimer) {
        window.clearTimeout(globalState.idleTimer);
        globalState.idleTimer = null;
      }
      globalState.refs += 1;
      return globalState.socket;
    }
  }

  const socket = io(endpoint, {
    transports: ["websocket"],
    reconnection: true,
    reconnectionAttempts: Infinity,
    reconnectionDelay: 750,
    reconnectionDelayMax: 5000,
    timeout: 20000,
  });
  socket.on("connect_error", (error) => {
    console.warn("[SOCKET] Connection error:", error.message);
  });
  socket.on("reconnect_attempt", (attempt) => {
    if (attempt % 10 === 0) {
      console.info("[SOCKET] Reconnect attempt", attempt);
    }
  });
  globalState.socket = socket;
  globalState.endpoint = endpoint;
  globalState.refs = 1;
  if (globalState.idleTimer) {
    window.clearTimeout(globalState.idleTimer);
    globalState.idleTimer = null;
  }
  return socket;
}

export function releaseSocket() {
  if (!globalState.socket) {
    return;
  }
  globalState.refs = Math.max(0, globalState.refs - 1);
  if (globalState.refs === 0) {
    scheduleIdleDisconnect();
  }
}
