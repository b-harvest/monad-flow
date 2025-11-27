import {
  BadRequestException,
  Injectable,
  Logger,
  NotFoundException,
} from '@nestjs/common';
import { InjectModel } from '@nestjs/mongoose';
import { Document, Model } from 'mongoose';
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

  async getAppMessage(id: string): Promise<any> {
    const message = await this.outboundRouterModel.findById(id).lean().exec();
    if (!message) {
      throw new NotFoundException(`Message with ID ${id} not found`);
    }
    return message;
  }

  async getLogsByTimeRange(from: Date, to: Date, type: string): Promise<any> {
    const query = {
      timestamp: { $gte: from, $lte: to },
    };

    switch (type) {
      case 'chunk':
        return this.chunkModel.find(query).lean().exec();
      case 'router':
        return this.outboundRouterModel.find(query).lean().exec();

      case 'offcpu':
        return this.offCpuModel.find(query).lean().exec();
      case 'scheduler':
        return this.schedulerModel.find(query).lean().exec();
      case 'perf':
        return this.perfStatModel.find(query).lean().exec();
      case 'turbo':
        return this.turboStatModel.find(query).lean().exec();
      case 'bpf':
        return this.bpfTraceModel.find(query).lean().exec();
      case 'bft':
        return this.bftLogModel.find(query).lean().exec();
      case 'exec':
        return this.execLogModel.find(query).lean().exec();

      default:
        throw new BadRequestException(
          `Invalid log type: ${type}. Available types: chunk, router, offcpu, scheduler, perf, turbo, bpf, bft, exec`,
        );
    }
  }

  async createFromUDP(payload: {
    type: NetworkEventType;
    data: any;
    timestamp: number;
    appMessageHash?: string;
  }): Promise<any> {
    const { type, data, timestamp, appMessageHash } = payload;
    if (type === NetworkEvent.MONAD_CHUNK) {
      return this.handleMonadChunkPacket(data, timestamp);
    } else if (type === NetworkEvent.OUTBOUND_ROUTER) {
      return this.handleOutboundRouter(data, timestamp, appMessageHash);
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
      this.persistDocument(doc);
      return doc;
    }

    if (isExec) {
      const doc = new this.execLogModel(payload);
      this.persistDocument(doc);
      return doc;
    }

    this.logger.warn(`[SYSTEM_LOG] Unknown unit=${unit}`);
    return null;
  }

  async saveBpfTrace(data: any) {
    this.logger.log(`[DB] Saving BpfTrace(Hook) func_name=${data.func_name}`);
    const doc = new this.bpfTraceModel({
      event_type: data.event_type,
      func_name: data.func_name,
      pid: data.pid,
      timestamp: data.timestamp,
      duration_ns: data.duration_ns,
      data: data.data,
    });

    this.persistDocument(doc);
    return doc;
  }

  async saveOffCpuEvent(data: any) {
    this.logger.log(`[DB] Saving OffCpu pid=${data.pid}`);
    const doc = new this.offCpuModel({
      timestamp: new Date(data.timestamp),
      process_name: data.process_name,
      pid: data.pid,
      tid: data.tid,
      duration_us: data.duration_us,
      stack: data.stack,
    });
    this.persistDocument(doc);
    return doc;
  }

  async saveSchedulerEvent(data: any) {
    this.logger.log(`[DB] Saving Scheduler pid=${data.main_pid}`);
    const doc = new this.schedulerModel({
      timestamp: new Date(data.timestamp),
      main_pid: data.main_pid,
      tid: data.tid,
      thread_name: data.thread_name,
      wait_delta_ms: data.wait_delta_ms,
      run_delta_ms: data.run_delta_ms,
      ctx_switches: data.ctx_switches,
    });
    this.persistDocument(doc);
    return doc;
  }

  async savePerfStatEvent(data: any) {
    this.logger.log(`[DB] Saving PerfStat pid=${data.pid}`);
    const doc = new this.perfStatModel({
      timestamp: new Date(data.timestamp),
      perf_timestamp: parseFloat(data.perf_timestamp),
      pid: data.pid,
      metrics: data.metrics,
    });
    this.persistDocument(doc);
    return doc;
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
    this.persistDocument(doc);
    return doc;
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
    this.persistDocument(doc);
    return doc;
  }

  private async handleOutboundRouter(
    data: any,
    timestamp: number,
    appMessageHash?: string,
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
      appMessageHash: appMessageHash,
      timestamp: new Date(timestamp / 1000),
    });
    this.persistDocument(doc);
    return doc;
  }

  private persistDocument(doc: Document) {
    doc.save().catch((error) => {
      this.logger.error(
        `Failed to persist document (${doc.constructor?.name ?? 'UnknownModel'})`,
        error.stack,
      );
    });
  }
}
