# API Contracts

Frontend API specification for Monad Flow monitoring system.

## Architecture Overview

The Monad Flow system uses a hybrid communication approach between Go probes and the NestJS backend:

### HTTP (Heavy Data)
- **Purpose**: Large RLP-encoded messages (OutboundRouterMessage)
- **Endpoint**: `POST /api/outbound-message`
- **Use Case**: Full consensus messages, state sync, peer discovery with complete payloads

### WebSocket (Light Real-time Data)
- **Purpose**: Real-time monitoring events and lightweight packet capture
- **Events**: Network packets, system metrics, performance counters
- **Use Case**: Live system monitoring, performance tracking, event tracing

### Data Flow
```
Go Probe (eBPF/Frida) → NestJS Backend → Frontend
                      ↓
                   MongoDB
```

---

## Common Type Definitions

These base types are used throughout the API:

```typescript
// Network layer identifiers (32-byte hashes)
type BlockID = string; // Hex-encoded hash with 0x prefix
type ConsensusBlockBodyId = string; // Hex-encoded hash with 0x prefix

// Numeric identifiers (uint64 in Go, string in TS for safety)
type Round = string;
type Epoch = string;
type SeqNum = string;

// Binary data (hex-encoded)
type NodeID = string; // Hex-encoded bytes
type Signature = string; // Hex-encoded bytes (65 bytes for ECDSA)
type RoundSignature = string; // Hex-encoded bytes
type BlsAggregateSignature = string; // Hex-encoded BLS signature

// Ethereum types
type Address = string; // Hex-encoded 20-byte address with 0x prefix
type Hash = string; // Hex-encoded 32-byte hash with 0x prefix
```

---

## HTTP API

### POST /api/outbound-message

Send heavy RLP-encoded Monad protocol messages.

**Request Body:**
```typescript
interface OutboundMessageRequest {
  type: string; // "OUTBOUND_ROUTER"
  appMessageHash: string; // Hex-encoded hash
  data: OutboundRouterCombined;
  timestamp: number; // Unix microseconds
}
```

#### OutboundRouterCombined

```typescript
interface OutboundRouterCombined {
  version: NetworkMessageVersion;
  messageType: number; // 1=PeerDiscovery, 2=FullNodesGroup, 3=AppMessage
  peerDiscovery?: PeerDiscoveryMessage;
  fullNodesGroup?: FullNodesGroupMessage;
  appMessage?: MonadMessage;
}

### GET /api/app-message/:id

Fetch a stored outbound app message by its MongoDB/DB identifier. Useful for drilling into proposal payloads referenced in other logs.

**Path Params**
- `id` (string): Document ID returned by previous API calls or WebSocket events.

**Response Body**
```json
{
  "_id": "675f4f...",
  "type": "OUTBOUND_ROUTER",
  "appMessageHash": "0x...",
  "data": { /* OutboundRouterCombined */ },
  "timestamp": 1732468800000
}
```

### GET /api/logs/:type

Query persisted monitoring logs within a time range. Historical playback (Forensics Mode) consumes this endpoint instead of a new `/api/visualizer/snapshots` API.

**Path Params**
- `type` (string): Log bucket to query. Supported string literals:
  - `chunk`: UDP chunk captures (MONAD_CHUNK_EVENT)
  - `router`: Outbound router events
  - `offcpu`: Off-CPU traces
  - `scheduler`: Scheduler stats
  - `perf`: Perf stat metrics
  - `turbo`: Turbo stat metrics
  - `bpf`: BPF trace events
  - `bft`: BFT/system logs
  - `exec`: Execution logs

**Query Params**
- `from` (ISO string): Start timestamp (inclusive).
- `to` (ISO string): End timestamp (inclusive).

**Response**
```json
[
  {
    "_id": "675f4f0a3b8e9d0012c9f111",
    "eventType": "MONAD_CHUNK_EVENT",
    "timestamp": "2024-11-24T12:00:00.123Z",   // ISO string
    "timestampMs": "1732468800123",            // where original payload includes it
    "payload": { /* identical to the WebSocket event payload */ }
  }
]
```
> 모든 로그 항목은 웹소켓으로 전달되는 구조와 1:1 매칭된다. 예를 들어 `chunk` 타입은 `MonadChunkPacketEvent`, `offcpu`는 `OffCpuEvent`, `perf`는 `PerfStatEvent` 그대로다. `router` 타입 역시 `/api/outbound-message`에서 반환되는 `OutboundRouterCombined` 포맷(또는 요약)과 동일하니, 프론트 타입 정의를 재사용하면 된다. `_id`는 MongoDB ObjectId 문자열이며, `timestamp`는 ISO 문자열, `timestampMs`는 원본 이벤트에 존재할 때만 문자열 형태로 포함된다.

> Historical Forensics Mode batches these responses per node/metric and caches them via the browser (IndexedDB/Cache Storage) for offline playback.

### Notes on WebSocket Bridge

The backend broadcasts to Socket.IO clients whenever a new outbound message is persisted. Endpoint `/api/outbound-message` emits `OUTBOUND_ROUTER_EVENT` with either the full object or (for proposals) a summary excluding `data`.

interface NetworkMessageVersion {
  serializeVersion: number; // uint32
  compressionVersion: number; // uint8
}

---

### Peer Discovery Messages

```typescript
interface PeerDiscoveryMessage {
  version: number; // uint16
  type: number; // 1=Ping, 2=Pong, 3=PeerLookupReq, 4=PeerLookupResp, 5=FullNodeRaptorcastReq, 6=FullNodeRaptorcastResp
  payload: Ping | Pong | PeerLookupRequest | PeerLookupResponse | FullNodeRaptorcastRequest | FullNodeRaptorcastResponse;
}

interface Ping {
  id: number; // uint32
  localNameRecord: MonadNameRecord | null;
}

interface Pong {
  pingID: number; // uint32
  localRecordSeq: string; // uint64
}

interface PeerLookupRequest {
  lookupID: number; // uint32
  target: NodeID;
  openDiscovery: boolean;
}

interface PeerLookupResponse {
  lookupID: number; // uint32
  target: NodeID;
  nameRecords: MonadNameRecord[];
}

interface FullNodeRaptorcastRequest {}

interface FullNodeRaptorcastResponse {}

interface MonadNameRecord {
  nameRecord: NameRecord;
  signature: Signature;
}

interface NameRecord {
  address: string; // IPv4 address string
  port: number; // uint16
  seq: string; // uint64
}
```

---

### Full Nodes Group Messages

```typescript
interface FullNodesGroupMessage {
  version: number; // uint8
  type: number; // 1=PrepareGroup, 2=PrepareGroupResponse, 3=ConfirmGroup
  payload: PrepareGroup | PrepareGroupResponse | ConfirmGroup;
}

interface PrepareGroup {
  validatorID: NodeID;
  maxGroupSize: string; // uint64
  startRound: Round;
  endRound: Round;
}

interface PrepareGroupResponse {
  req: PrepareGroup;
  nodeID: NodeID;
  accept: boolean;
}

interface ConfirmGroup {
  prepare: PrepareGroup;
  peers: NodeID[];
  nameRecords: MonadNameRecord[];
}
```

---

### Monad Application Messages

```typescript
interface MonadMessage {
  version: MonadVersion;
  typeId: number; // 1=Consensus, 2=ForwardedTx, 3=StateSync, 4=BlockSyncRequest, 5=BlockSyncResponse
  payload: ConsensusMessage | ForwardedTxMessage | StateSyncNetworkMessage | BlockSyncRequest | BlockSyncResponse;
}

interface MonadVersion {
  protocolVersion: number; // uint32
  clientVersionMajor: number; // uint16
  clientVersionMinor: number; // uint16
  hashVersion: number; // uint16
  serializeVersion: number; // uint16
}
```

---

### Consensus Messages

```typescript
interface ConsensusMessage {
  version: number; // uint32
  payload: ProtocolMessage;
}

interface ProtocolMessage {
  name: string; // "ProtocolMessage"
  messageType: number; // 1=Proposal, 2=Vote, 3=Timeout, 4=RoundRecovery, 5=NoEndorsement, 6=AdvanceRound
  payload: ProposalMessage | VoteMessage | TimeoutMessage | RoundRecoveryMessage | NoEndorsementMessage | AdvanceRoundMessage;
}
```

#### Proposal Message

```typescript
interface ProposalMessage {
  proposalRound: Round;
  proposalEpoch: Epoch;
  tip: ConsensusTip;
  blockBody: ConsensusBlockBody;
  lastRoundTC?: TimeoutCertificate | null;
}

interface ConsensusTip {
  blockHeader: ConsensusBlockHeader;
  signature: string; // Hex-encoded bytes
  freshCertificate?: FreshProposalCertificateWrapper | null;
}

interface ConsensusBlockHeader {
  blockRound: Round;
  epoch: Epoch;
  qc: QuorumCertificate;
  author: NodeID;
  seqNum: SeqNum;
  timestampNS: string; // BigInt as string
  roundSignature: RoundSignature;
  delayedExecutionResults: FinalizedHeader[];
  executionInputs: ProposedHeader;
  blockBodyID: ConsensusBlockBodyId;
  baseFee?: string | null; // Optional uint64
  baseFeeTrend?: string | null; // Optional uint64
  baseFeeMoment?: string | null; // Optional uint64
}

interface ProposedHeader {
  ommersHash: Hash;
  beneficiary: Address;
  transactionsRoot: Hash;
  difficulty: string; // uint64
  number: string; // uint64
  gasLimit: string; // uint64
  timestamp: string; // uint64
  extraData: string; // 32 bytes hex-encoded
  mixHash: Hash;
  nonce: string; // 8 bytes hex-encoded
  baseFeePerGas: string; // uint64
  withdrawalsRoot: Hash;
  blobGasUsed: string; // uint64
  excessBlobGas: string; // uint64
  parentBeaconBlockRoot: Hash;
  requestsHash?: Hash | null;
}

interface FinalizedHeader {
  // Ethereum types.Header structure from go-ethereum
  [key: string]: any;
}

interface ConsensusBlockBody {
  executionBody: ExecutionBody;
}

interface ExecutionBody {
  transactions: Transaction[];
  ommers: Ommer[];
  withdrawals: Withdrawal[];
}

interface Transaction {
  // Ethereum transaction structure from go-ethereum
  [key: string]: any;
}

interface Ommer {}

interface Withdrawal {
  // Ethereum withdrawal structure
  [key: string]: any;
}

interface QuorumCertificate {
  info: Vote;
  signatures: string; // RLP raw value (hex-encoded)
}

interface Vote {
  id: BlockID;
  round: Round;
  epoch: Epoch;
}

interface FreshProposalCertificateWrapper {
  typeId: number; // 1=NEC, 2=NoTip
  certificate: FreshProposalCertificateNEC | FreshProposalCertificateNoTip;
}

interface FreshProposalCertificateNEC {
  nec: NoEndorsementCertificate;
}

interface FreshProposalCertificateNoTip {
  noTip: NoTipCertificate;
}

interface NoEndorsementCertificate {
  msg: NoEndorsement;
  signatures: string; // Hex-encoded bytes
}

interface NoEndorsement {
  epoch: Epoch;
  round: Round;
  tipQCRound: Round;
}

interface NoTipCertificate {
  epoch: Epoch;
  round: Round;
  tipRounds: HighTipRoundSigColTuple[];
  highQc: QuorumCertificate;
}

interface TimeoutCertificate {
  epoch: Epoch;
  round: Round;
  tipRounds: HighTipRoundSigColTuple[];
  highExtend?: HighExtendWrapper | null;
}

interface HighTipRoundSigColTuple {
  highQCRound: Round;
  highTipRound: Round;
  sigs: SignatureCollection;
}

interface SignatureCollection {
  signers: SignerMap;
  sig: BlsAggregateSignature;
}

interface SignerMap {
  numBits: number; // uint32
  buf: string; // Hex-encoded bytes
}

interface HighExtendWrapper {
  typeId: number; // 1=Tip, 2=Qc
  extend: HighExtendTip | HighExtendQc;
}

interface HighExtendTip {
  tip: ConsensusTip;
  voteSignature: string; // Hex-encoded bytes
}

interface HighExtendQc {
  qc: QuorumCertificate;
}
```

#### Vote Message

```typescript
interface VoteMessage {
  vote: Vote;
  sig: string; // Hex-encoded signature bytes
}
```

#### Timeout Message

```typescript
interface TimeoutMessage {
  tmInfo: TimeoutInfo;
  timeoutSignature: string; // Hex-encoded bytes
  highExtend: HighExtendWrapper;
  lastRoundCertificate?: RoundCertificateWrapper | null;
}

interface TimeoutInfo {
  epoch: Epoch;
  round: Round;
  highQCRound: Round;
  highTipRound: Round;
}

interface RoundCertificateWrapper {
  typeId: number; // 1=QC, 2=TC
  certificate: RoundCertificateQC | RoundCertificateTC;
}

interface RoundCertificateQC {
  qc: QuorumCertificate;
}

interface RoundCertificateTC {
  tc: TimeoutCertificate;
}
```

#### Round Recovery Message

```typescript
interface RoundRecoveryMessage {
  round: Round;
  epoch: Epoch;
  tc: TimeoutCertificate;
}
```

#### No Endorsement Message

```typescript
interface NoEndorsementMessage {
  msg: NoEndorsement;
  sig: string; // Hex-encoded bytes
}
```

#### Advance Round Message

```typescript
interface AdvanceRoundMessage {
  lastRoundCertificate: RoundCertificateWrapper;
}
```

---

### Forwarded Transaction Messages

```typescript
type ForwardedTxMessage = string[][]; // Array of hex-encoded transaction bytes
```

---

### State Sync Messages

```typescript
interface StateSyncNetworkMessage {
  messageName: string; // "StateSync"
  typeID: number; // 1=Request, 2=Response, 3=BadVersion, 4=Completion
  request?: StateSyncRequest;
  response?: StateSyncResponse;
  badVersion?: StateSyncBadVersion;
  completion?: SessionId;
}

interface SessionId {
  value: string; // uint64
}

interface StateSyncRequest {
  version: StateSyncVersion;
  prefix: string; // uint64
  prefixBytes: number; // uint8
  target: string; // uint64
  from: string; // uint64
  until: string; // uint64
  oldTarget: string; // uint64
}

interface StateSyncResponse {
  version: StateSyncVersion;
  nonce: string; // uint64
  responseIndex: number; // uint32
  request: StateSyncRequest;
  response: StateSyncUpsertV1[];
  responseN: string; // uint64
}

interface StateSyncUpsertV1 {
  upsertType: StateSyncUpsertType;
  data: string; // Hex-encoded bytes
}

type StateSyncUpsertType = 1 | 2 | 3 | 4 | 5 | 6;
// 1=Code, 2=Account, 3=Storage, 4=AccountDelete, 5=StorageDelete, 6=Header

interface StateSyncVersion {
  major: number; // uint16
  minor: number; // uint16
}

interface StateSyncBadVersion {
  minVersion: StateSyncVersion;
  maxVersion: StateSyncVersion;
}
```

---

### Block Sync Messages

```typescript
interface BlockSyncRequest {
  messageName: string; // "BlockSyncRequest"
  typeId: number; // 1=Headers, 2=Body
  isHeaders?: boolean;
  headers?: BlockRange;
  isPayload?: boolean;
  payload?: ConsensusBlockBodyId;
}

interface BlockRange {
  lastBlockID: BlockID;
  numBlocks: SeqNum;
}

interface BlockSyncResponse {
  messageName: string; // "BlockSyncResponse"
  typeId: number; // 1=Headers, 2=Body
  headersData?: BlockSyncHeadersResponse;
  payloadData?: BlockSyncBodyResponse;
}

interface BlockSyncHeadersResponse {
  messageName: string; // "BlockSyncHeadersResponse"
  typeId: number; // 1=Found, 2=NotAvailable
  foundRange?: BlockRange;
  foundHeaders?: ConsensusBlockHeader[];
  notAvailRange?: BlockRange;
}

interface BlockSyncBodyResponse {
  messageName: string; // "BlockSyncBodyResponse"
  typeId: number; // 1=Found, 2=NotAvailable
  foundBody?: ConsensusBlockBody;
  notAvailPayload?: ConsensusBlockBodyId;
}
```

---

## WebSocket Events

The backend broadcasts these events to connected clients via WebSocket (Socket.IO). Subscribe to all channels that matter to your UI.

### Event: `OUTBOUND_ROUTER_EVENT`

Emitted whenever `POST /api/outbound-message` persists a record. Proposals (`messageType === 1`) omit the heavy `data` payload to reduce bandwidth.

```typescript
interface OutboundRouterEvent {
  _id: string;
  type: string;
  messageType: number;
  appMessageHash?: string;
  timestamp: number; // unix ms
  data?: OutboundRouterCombined; // only present when messageType !== 1
}
```

### Event: `MONAD_CHUNK_EVENT`

UDP packet capture of Monad chunk packets.

```typescript
interface MonadChunkPacketEvent {
  network: {
    ipv4: {
      srcIp: string;
      dstIp: string;
      protocol: string;
    };
    port: {
      srcPort: number;
      dstPort: number;
    };
  };
  signature: string; // 65 bytes hex-encoded
  version: number; // uint16
  flags: number; // uint8
  broadcast: boolean;
  secondaryBroadcast: boolean;
  merkleTreeDepth: number; // uint8 (4 bits from flags)
  epoch: string; // uint64
  timestampMs: string; // uint64
  appMessageHash: string; // 20 bytes hex-encoded
  appMessageLen: number; // uint32
  merkleProof: string[][]; // Array of 20-byte hex arrays
  firstHopRecipient: string; // 20 bytes hex-encoded
  merkleLeafIdx: number; // uint8
  reserved: number; // uint8
  chunkId: number; // uint16
  timestamp: string; // ISO 8601 timestamp
}
```

**Example Payload:**
```json
{
  "network": {
    "ipv4": {
      "srcIp": "10.0.1.100",
      "dstIp": "10.0.1.200",
      "protocol": "UDP"
    },
    "port": {
      "srcPort": 45678,
      "dstPort": 8000
    }
  },
  "signature": "0x1234567890abcdef...",
  "version": 1,
  "flags": 192,
  "broadcast": true,
  "secondaryBroadcast": false,
  "merkleTreeDepth": 4,
  "epoch": "12345",
  "timestampMs": "1732468800000",
  "appMessageHash": "0xabcdef1234567890...",
  "appMessageLen": 2048,
  "merkleProof": [
    ["0x1111...", "0x2222..."],
    ["0x3333...", "0x4444..."]
  ],
  "firstHopRecipient": "0x9876543210abcdef...",
  "merkleLeafIdx": 2,
  "reserved": 0,
  "chunkId": 42,
  "timestamp": "2024-11-24T12:00:00.000Z"
}
```

---

### Event: `BPF_TRACE`

Function hooking events via Frida (Enter/Exit).

```typescript
interface BpfTraceEvent {
  event_type: "enter" | "exit";
  func_name: string; // Mangled C++ function name (truncated to 70 chars)
  pid: string;
  timestamp: string; // ISO 8601
  duration_ns: string; // "0" for enter, actual duration for exit
  data: EnterData | ExitData;
}

interface EnterData {
  caller_name: string; // Function that called this function
  args_hex: string[]; // First 5 arguments as hex strings
}

interface ExitData {
  back_to_name: string; // Return destination
  return_value: string; // Hex-encoded return value
}
```

**Example Payload (Enter):**
```json
{
  "event_type": "enter",
  "func_name": "_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision...",
  "pid": "12345",
  "timestamp": "2024-11-24T12:00:00.123Z",
  "duration_ns": "0",
  "data": {
    "caller_name": "main+0x1234",
    "args_hex": ["0x7fff1234", "0x5678", "0x0", "0x0", "0x0"]
  }
}
```

**Example Payload (Exit):**
```json
{
  "event_type": "exit",
  "func_name": "_ZN5monad13execute_blockINS_11MonadTraitsIL14monad_revision...",
  "pid": "12345",
  "timestamp": "2024-11-24T12:00:01.234Z",
  "duration_ns": "1234567890",
  "data": {
    "back_to_name": "caller",
    "return_value": "0x1"
  }
}
```

---

### Event: `SYSTEM_LOG`

Systemd journal logs from monitored services.

```typescript
interface SystemLogEvent {
  service: string; // systemd unit name (e.g., "monad-node.service")
  timestamp: string; // Format: "2006-01-02 15:04:05.000000"
  message: string; // Log message content
  pid?: string; // Process ID (optional)
}
```

**Example Payload:**
```json
{
  "service": "monad-bft.service",
  "timestamp": "2024-11-24 12:00:00.123456",
  "message": "Consensus round 12345 completed successfully",
  "pid": "67890"
}
```

---

### Event: `OFF_CPU`

Off-CPU time tracking (thread blocked/waiting).

```typescript
interface OffCpuEvent {
  timestamp: string; // Format: "2006-01-02 15:04:05.000000"
  process_name: string;
  pid: string;
  tid: string; // Thread ID
  duration_us: number; // Microseconds off-CPU
  stack: string[]; // Stack trace lines
}
```

**Example Payload:**
```json
{
  "timestamp": "2024-11-24 12:00:00.123456",
  "process_name": "monad-execution",
  "pid": "12345",
  "tid": "12347",
  "duration_us": 5432,
  "stack": [
    "futex_wait_queue",
    "futex_wait",
    "__pthread_mutex_lock",
    "monad::StateDB::commit"
  ]
}
```

---

### Event: `SCHEDULER`

Thread scheduler statistics (latency and context switches).

```typescript
interface SchedulerEvent {
  timestamp: string; // Format: "2006-01-02 15:04:05.000000"
  main_pid: string; // Main process PID
  tid: string; // Thread ID
  thread_name: string; // Thread name from /proc/[pid]/task/[tid]/comm
  wait_delta_ms: number; // Wait time delta (milliseconds)
  run_delta_ms: number; // Run time delta (milliseconds)
  ctx_switches: string; // Context switch count delta (uint64)
}
```

**Example Payload:**
```json
{
  "timestamp": "2024-11-24 12:00:01.000000",
  "main_pid": "12345",
  "tid": "12347",
  "thread_name": "consensus_thd",
  "wait_delta_ms": 12.5,
  "run_delta_ms": 987.5,
  "ctx_switches": "42"
}
```

---

### Event: `PERF_STAT`

Performance counter statistics (perf stat).

```typescript
interface PerfStatEvent {
  timestamp: string; // Format: "2006-01-02 15:04:05.000000"
  perf_timestamp: string; // Perf's own timestamp (float as string)
  pid: string;
  metrics: PerfMetric[];
}

interface PerfMetric {
  event: string; // Event name (e.g., "cycles", "instructions", "cache-misses")
  value: string; // Counter value or "Not Supported"
  unit?: string; // Unit (e.g., "msec")
  metric_val?: string; // Derived metric (e.g., "2.5 GHz", "0.85 IPC")
  run_pct?: string; // Run percentage (e.g., "100.00%")
}
```

**Example Payload:**
```json
{
  "timestamp": "2024-11-24 12:00:01.000000",
  "perf_timestamp": "1.234567",
  "pid": "12345",
  "metrics": [
    {
      "event": "cycles",
      "value": "1234567890",
      "metric_val": "2.47 GHz",
      "run_pct": "100.00%"
    },
    {
      "event": "instructions",
      "value": "987654321",
      "metric_val": "0.80 insn per cycle",
      "run_pct": "100.00%"
    },
    {
      "event": "cache-misses",
      "value": "12345",
      "metric_val": "5.2% of all cache refs",
      "run_pct": "100.00%"
    }
  ]
}
```

---

### Event: `TURBO_STAT`

CPU frequency and power consumption metrics.

```typescript
interface TurboStatEvent {
  timestamp: string; // Format: "2006-01-02 15:04:05.000"
  core: string; // Core number or "-" for package
  cpu: string; // CPU number or "-" for aggregated
  avg_mhz: number; // Average MHz
  busy_pct: number; // CPU busy percentage
  bzy_mhz: number; // Busy MHz
  tsc_mhz: number; // TSC MHz
  ipc: number; // Instructions per cycle
  irq: number; // IRQ count
  cor_watt: number; // Core power (watts)
  pkg_watt: number; // Package power (watts)
}
```

**Example Payload:**
```json
{
  "timestamp": "2024-11-24 12:00:01.000",
  "core": "0",
  "cpu": "0",
  "avg_mhz": 2400.5,
  "busy_pct": 85.3,
  "bzy_mhz": 2815.2,
  "tsc_mhz": 2800.0,
  "ipc": 1.45,
  "irq": 12345,
  "cor_watt": 15.2,
  "pkg_watt": 45.8
}
```

---

## Event Naming Convention

All WebSocket events follow the enum definitions in `backend/src/common/enum-definition.ts`:

- `MONAD_CHUNK_EVENT`: Network packet capture
- `CLIENT_EVENT`: Broadcast to frontend clients
- `BPF_TRACE`: Function hooking traces
- `SYSTEM_LOG`: Systemd journal logs
- `OFF_CPU`: Off-CPU time events
- `SCHEDULER`: Scheduler statistics
- `PERF_STAT`: Performance counters
- `TURBO_STAT`: CPU power/frequency

---

## Type ID Reference

### FreshProposalCertificate Types
- `1`: NoEndorsementCertificate (NEC)
- `2`: NoTipCertificate

### HighExtend Types
- `1`: HighExtendTip
- `2`: HighExtendQc

### RoundCertificate Types
- `1`: QuorumCertificate (QC)
- `2`: TimeoutCertificate (TC)

### Protocol Message Types
- `1`: ProposalMessage
- `2`: VoteMessage
- `3`: TimeoutMessage
- `4`: RoundRecoveryMessage
- `5`: NoEndorsementMessage
- `6`: AdvanceRoundMessage

### Monad Message Types
- `1`: ConsensusMessage
- `2`: ForwardedTxMessage
- `3`: StateSyncMessage
- `4`: BlockSyncRequest
- `5`: BlockSyncResponse

### StateSync Types
- `1`: Request
- `2`: Response
- `3`: BadVersion
- `4`: Completion

### StateSyncUpsert Types
- `1`: Code
- `2`: Account
- `3`: Storage
- `4`: AccountDelete
- `5`: StorageDelete
- `6`: Header

### BlockSync Types
- `1`: Headers
- `2`: Body

### BlockSync Response Status
- `1`: Found
- `2`: NotAvailable

---

## Notes

### Type Safety

1. **Large Numbers**: Go `uint64` types are serialized as strings in JSON to avoid JavaScript's `Number.MAX_SAFE_INTEGER` (2^53-1) limitation.

2. **Binary Data**: All byte arrays (`[]byte`, `[N]byte`) are hex-encoded with `0x` prefix.

3. **Hashes**: Ethereum-style 32-byte hashes are represented as hex strings with `0x` prefix.

4. **Addresses**: 20-byte Ethereum addresses are hex strings with `0x` prefix.

5. **Timestamps**:
   - ISO 8601 format for absolute times: `"2024-11-24T12:00:00.000Z"`
   - Custom format for logs: `"2006-01-02 15:04:05.000000"`
   - Unix microseconds for network events: `number`

### Optional Fields

Fields marked with `?` or `| null` are optional and may not be present in the payload. Always check for existence before accessing.

### Discriminated Unions

Many interfaces use discriminated unions with `typeId` fields:
- Check `typeId` first to determine the actual type
- Type narrowing should be based on `typeId` value
- Example: `FreshProposalCertificateWrapper.typeId === 1` means `certificate` is `FreshProposalCertificateNEC`

### RLP Encoding

Messages sent via HTTP contain RLP-encoded data. The Go probe decodes RLP and sends structured JSON to the backend. Frontend receives fully decoded JSON objects.

### Real-time Updates

WebSocket events are broadcast immediately upon receipt. There is no buffering or batching. High-frequency events (e.g., PERF_STAT, TURBO_STAT) may require throttling on the frontend.

### Version Compatibility

- **MonadVersion**: Protocol, client, hash, and serialization versions
- **StateSyncVersion**: Major and minor version for state sync compatibility
- Always check version compatibility before processing messages
