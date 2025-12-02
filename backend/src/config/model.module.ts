import { Module } from '@nestjs/common';
import { MongooseModule } from '@nestjs/mongoose';
import {
  OutboundRouterMessage,
  OutboundRouterMessageSchema,
} from '../schema/network/outbound-router-message.schema';
import {
  MonadChunkPacket,
  MonadChunkPacketSchema,
} from '../schema/network/monad-chunk-packet.schema';
import {
  BpfTraceLog,
  BpfTraceLogSchema,
} from '../schema/system/bpf-trace-log.schema';
import {
  OffCpuLog,
  OffCpuLogSchema,
} from '../schema/system/off-cpu-log.schema';
import {
  SchedulerLog,
  SchedulerLogSchema,
} from '../schema/system/scheduler-log.schema';
import {
  PerfStatLog,
  PerfStatLogSchema,
} from '../schema/system/perfstat-log.schema';
import {
  TurboStatLog,
  TurboStatLogSchema,
} from '../schema/system/turbostat-log.schema';
import {
  MonadBftLog,
  MonadBftLogSchema,
} from '../schema/system/monad-bft-log.schema';
import {
  MonadExecutionLog,
  MonadExecutionLogSchema,
} from '../schema/system/monad-execution-log.schema';
import {
  PingLatency,
  PingLatencySchema,
} from 'src/schema/network/ping-latency.schema';
import { Leader, LeaderSchema } from 'src/schema/network/leader.schema';

@Module({
  imports: [
    MongooseModule.forFeature([
      {
        name: OutboundRouterMessage.name,
        schema: OutboundRouterMessageSchema,
        collection: 'outbound_router_messages',
      },
      {
        name: MonadChunkPacket.name,
        schema: MonadChunkPacketSchema,
        collection: 'monad_chunk_packets',
      },
      {
        name: BpfTraceLog.name,
        schema: BpfTraceLogSchema,
        collection: 'bpf_trace_events',
      },
      {
        name: OffCpuLog.name,
        schema: OffCpuLogSchema,
        collection: 'offcpu_events',
      },
      {
        name: SchedulerLog.name,
        schema: SchedulerLogSchema,
        collection: 'scheduler_events',
      },
      {
        name: PerfStatLog.name,
        schema: PerfStatLogSchema,
        collection: 'perf_stat_events',
      },
      {
        name: TurboStatLog.name,
        schema: TurboStatLogSchema,
        collection: 'turbostat_events',
      },
      {
        name: MonadBftLog.name,
        schema: MonadBftLogSchema,
        collection: 'monad_bft_logs',
      },
      {
        name: MonadExecutionLog.name,
        schema: MonadExecutionLogSchema,
        collection: 'monad_execution_logs',
      },
      {
        name: PingLatency.name,
        schema: PingLatencySchema,
        collection: 'ping_latencies',
      },
      {
        name: Leader.name,
        schema: LeaderSchema,
        collection: 'leaders',
      },
    ]),
  ],
  exports: [MongooseModule],
})
export class ModelModule {}
