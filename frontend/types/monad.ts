export type MonadNodeState = "leader" | "active" | "idle" | "failed" | "syncing";

export type RaptorGroup = "primary" | "secondary";

export interface MonadNode {
  id: string;
  name: string;
  role: "leader" | "validator";
  ip: string;
  uptimePct: number;
  participationRate: number;
  lastActivity: string;
  state: MonadNodeState;
  position: [number, number, number];
  cluster?: RaptorGroup;
  parentId?: string | null;
  isLocal?: boolean;
  latitude?: number;
  longitude?: number;
}

export interface BpfTraceSample {
  eventType: "enter" | "exit";
  funcName: string;
  timestamp: string;
  durationNs?: string;
  detail: string;
}

export interface SystemLogSample {
  service: string;
  timestamp: string;
  message: string;
}

export interface OffCpuSample {
  processName: string;
  durationUs: number;
  stack: string[];
}

export interface SchedulerSample {
  threadName: string;
  waitDeltaMs: number;
  runDeltaMs: number;
  ctxSwitches: string;
}

export interface PerfMetricSample {
  event: string;
  value: string;
  metricVal?: string;
  runPct?: string;
}

export interface TurboStatSample {
  core: string;
  avgMhz: number;
  busyPct: number;
  ipc: number;
  pkgWatt: number;
}

export interface NodeTelemetryDigest {
  bpfTrace: BpfTraceSample[];
  systemLogs: SystemLogSample[];
  offCpu: OffCpuSample;
  scheduler: SchedulerSample;
  perfStat: PerfMetricSample[];
  turboStat: TurboStatSample;
  updatedAt: number;
}

export interface ConsensusMetrics {
  epoch: number;
  round: number;
  leaderId: string;
  tps: number;
  blockHeight: number;
  avgBlockTime: number;
  networkHealth: number;
  connectionStatus: "connected" | "degraded" | "lost";
  timestamp: number;
}

export type MonitoringSeverity = "info" | "warning" | "critical";

export interface MonitoringEvent {
  id: string;
  timestamp: number;
  nodeId?: string;
  label: string;
  detail: string;
  severity: MonitoringSeverity;
}

export type PulseVisualType = "proposal" | "vote" | "pulse" | "chunk";

export interface PulseVisualEffect {
  id: string;
  type: PulseVisualType;
  fromNodeId: string;
  toNodeId?: string;
  createdAt: number;
  ttl: number;
  direction?: "inbound" | "outbound";
}

export interface PlaybackState {
  mode: "live" | "historical";
  range: {
    from: number;
    to: number;
  };
  cursor: number;
  isPlaying: boolean;
  speed: 0.25 | 0.5 | 1 | 2 | 4;
  liveAvailable: boolean;
}

export interface AlertToast {
  id: string;
  title: string;
  description: string;
  severity: MonitoringSeverity;
  createdAt: number;
}

export interface OutboundRouterEventSummary {
  id: string;
  messageType: number;
  appMessageHash?: string;
  timestamp: number;
  peer: string;
  size: number;
  status: "delivered" | "pending" | "blocked";
  detail: string;
}

export interface SocketEventRecord {
  event: string;
  payload: string;
  timestamp: number;
}

export interface ChunkPacketRecord {
  id: string;
  appMessageHash?: string;
  chunkId: number;
  timestamp: number;
  fromIp: string;
  fromPort: number;
  toIp: string;
  toPort: number;
  size: number;
  payload: unknown;
}
