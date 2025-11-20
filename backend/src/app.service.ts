import { Injectable, Logger } from '@nestjs/common';
import { InjectModel } from '@nestjs/mongoose';
import { Model } from 'mongoose';
import { OutboundRouterMessage } from './schema/network/outbound-router-message.schema';
import { NetworkEvent } from './common/enum-definition';
import { NetworkEventType } from './common/type-definition';
import { MonadChunkPacket } from './schema/network/monad-chunk-packet.schema';
import { OffCpuLog } from './schema/system/off-cpu-log.schema';
import { SchedulerLog } from './schema/system/scheduler-log.schema';
import { PerfStatLog } from './schema/system/perfstat-log.schema';
import { TurboStatLog } from './schema/system/turbostat-log.schema';
import { BpfTraceLog } from './schema/system/bpf-trace-log.schema';
import { MonadExecutionLog } from './schema/system/monad-execution-log.schema';
import { MonadBftLog } from './schema/system/monad-bft-log.schema';

@Injectable()
export class AppService {
  private readonly logger = new Logger(AppService.name);

  constructor(
    @InjectModel(MonadChunkPacket.name)
    private readonly chunkModel: Model<MonadChunkPacket>,
    @InjectModel(OutboundRouterMessage.name)
    private readonly outboundRouterModel: Model<OutboundRouterMessage>,
    @InjectModel(OffCpuLog.name)
    private readonly offCpuModel: Model<OffCpuLog>,
    @InjectModel(SchedulerLog.name)
    private readonly schedulerModel: Model<SchedulerLog>,
    @InjectModel(PerfStatLog.name)
    private readonly perfStatModel: Model<PerfStatLog>,
    @InjectModel(TurboStatLog.name)
    private readonly turboStatModel: Model<TurboStatLog>,
    @InjectModel(BpfTraceLog.name)
    private readonly bpfTraceModel: Model<BpfTraceLog>,
    @InjectModel(MonadBftLog.name)
    private readonly bftLogModel: Model<MonadBftLog>,
    @InjectModel(MonadExecutionLog.name)
    private readonly execLogModel: Model<MonadExecutionLog>,
  ) {}

  async getAll(): Promise<void> {}

  async createFromUDP(payload: {
    type: NetworkEventType;
    data: any;
    timestamp: number;
  }): Promise<any> {
    const { type, data, timestamp } = payload;
    if (type === NetworkEvent.MONAD_CHUNK) {
      return this.handleMonadChunkPacket(data, timestamp);
    } else if (type === NetworkEvent.OUTBOUND_ROUTER) {
      return this.handleOutboundRouter(data, timestamp);
    } else {
      this.logger.log(
        `Received UDP event:\n${JSON.stringify(payload, null, 2)}`,
      );
    }
  }

  async saveSystemdLog(data: any) {
    this.logger.log(`[DB] Saving SystemLog (${data.service})`);
    const unit = data.service || '';
    const timestamp = new Date(data.timestamp);

    const isBft = unit.includes('bft');
    const isExec = unit.includes('execution');

    const payload = {
      unit,
      message: data.message,
      timestamp,
      pid: data.pid,
    };

    if (isBft) {
      const doc = new this.bftLogModel(payload);
      return doc.save();
    }

    if (isExec) {
      const doc = new this.execLogModel(payload);
      return doc.save();
    }

    this.logger.warn(`[SYSTEM_LOG] Unknown unit=${unit}`);
    return null;
  }

  async saveBpfTrace(data: any) {
    this.logger.log(`[DB] Saving BpfTrace(Hook)`);
    const doc = new this.bpfTraceModel({
      event_type: data.event_type,
      timestamp: this.parsedTimestamp(data.timestamp),
      pid: data.pid,
      tid: data.tid,
      data: data.data,
    });

    return doc.save();
  }

  async saveOffCpuEvent(data: any) {
    this.logger.log(`[DB] Saving OffCpu`);
    const doc = new this.offCpuModel({
      timestamp: new Date(data.timestamp),
      process_name: data.process_name,
      tid: data.tid,
      duration_us: data.duration_us,
      stack: data.stack,
    });
    return doc.save();
  }

  async saveSchedulerEvent(data: any) {
    this.logger.log(`[DB] Saving Scheduler`);
    const doc = new this.schedulerModel({
      timestamp: new Date(data.timestamp),
      main_pid: data.main_pid,
      tid: data.tid,
      thread_name: data.thread_name,
      wait_delta_ms: data.wait_delta_ms,
      run_delta_ms: data.run_delta_ms,
      ctx_switches: data.ctx_switches,
    });
    return doc.save();
  }

  async savePerfStatEvent(data: any) {
    this.logger.log(`[DB] Saving PerfStat`);
    const doc = new this.perfStatModel({
      timestamp: new Date(data.timestamp),
      perf_timestamp: parseFloat(data.perf_timestamp),
      pid: data.pid,
      metrics: data.metrics,
    });
    return doc.save();
  }

  async saveTurboStatEvent(data: any) {
    this.logger.log(`[DB] Saving TurboStat`);
    const doc = new this.turboStatModel({
      timestamp: new Date(data.timestamp),
      core: data.core,
      cpu: data.cpu,
      avg_mhz: data.avg_mhz,
      busy_pct: data.busy_pct,
      bzy_mhz: data.bzy_mhz,
      tsc_mhz: data.tsc_mhz,
      ipc: data.ipc,
      irq: data.irq,
      cor_watt: data.cor_watt,
      pkg_watt: data.pkg_watt ?? null,
    });
    return doc.save();
  }

  private async handleMonadChunkPacket(
    data: any,
    timestamp: number,
  ): Promise<MonadChunkPacket> {
    this.logger.log(
      `[DB] Saving MonadChunkPacket epoch=${data.Epoch}, chunk=${data.ChunkID}, appMessageHash=${data.AppMessageHash}`,
    );
    const doc = new this.chunkModel({
      network: {
        ipv4: {
          srcIp: data.Network?.Ipv4?.SrcIp,
          dstIp: data.Network?.Ipv4?.DstIp,
          protocol: data.Network?.Ipv4?.Protocol,
        },
        port: {
          srcPort: data.Network?.Port?.SrcPort,
          dstPort: data.Network?.Port?.DstPort,
        },
      },

      signature: '0x' + Buffer.from(data.Signature || []).toString('hex'),
      version: data.Version,
      flags: data.Flags,
      broadCast: data.Broadcast,
      secondaryBroadcast: data.SecondaryBroadcast,
      merkleTreeDepth: data.MerkleTreeDepth,

      epoch: data.Epoch?.toString(),
      timestampMs: data.TimestampMs,

      appMessageHash:
        '0x' + Buffer.from(data.AppMessageHash || []).toString('hex'),
      appMessageLen: data.AppMessageLen,

      merkleProof: data.MerkleProof,
      firstHopRecipient:
        '0x' + Buffer.from(data.FirstHopRecipient || []).toString('hex'),

      merkleLeafIdx: data.MerkleLeafIdx,
      reserved: data.Reserved,
      chunkId: data.ChunkID,

      timestamp: new Date(timestamp / 1000),
    });
    return doc.save();
  }

  private async handleOutboundRouter(
    data: any,
    timestamp: number,
  ): Promise<OutboundRouterMessage> {
    const jsonString = JSON.stringify(data);
    const sizeBytes = Buffer.byteLength(jsonString);
    const sizeKB = (sizeBytes / 1024).toFixed(2);

    const sizeDisplay =
      sizeBytes > 1024 * 1024
        ? `${(sizeBytes / (1024 * 1024)).toFixed(2)} MB`
        : `${sizeKB} KB`;

    this.logger.log(
      `[DB] Saving OutboundRouterMessage messageType=${data.messageType}, size=${sizeDisplay}`,
    );
    const doc = new this.outboundRouterModel({
      version: data.version,
      messageType: data.messageType,
      data:
        data.peerDiscovery || data.fullNodesGroup || data.appMessage || null,
      appMessageHash: data.appMessageHash,
      timestamp: new Date(timestamp / 1000),
    });
    return doc.save();
  }

  private parsedTimestamp(timestamp: string): Date {
    return new Date(
      timestamp.replace(' ', 'T').replace(/(\.\d{6})$/, (m) => {
        const ms = Math.floor(Number(m) * 1000);
        return `.${ms}Z`;
      }),
    );
  }
}
