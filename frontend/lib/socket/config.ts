export const SOCKET_EVENT_NAMES = [
  "MONAD_CHUNK",
  "OUTBOUND_ROUTER",
  "BPF_TRACE",
  "SYSTEM_LOG",
  "OFF_CPU",
  "SCHEDULER",
  "PERF_STAT",
  "TURBO_STAT",
  "BFT",
  "LEADER",
];

export const defaultSocketEndpoint =
  process.env.BACKEND_SOCKET_URL ?? "http://51.195.24.236:3000";
