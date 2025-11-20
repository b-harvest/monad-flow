import {
  WebSocketGateway,
  WebSocketServer,
  OnGatewayConnection,
  OnGatewayDisconnect,
  SubscribeMessage,
  MessageBody,
  OnGatewayInit,
} from '@nestjs/websockets';
import { Server, Socket } from 'socket.io';
import { Logger } from '@nestjs/common';
import { WebsocketEventType } from '../common/type-definition';
import { WebsocketEvent } from '../common/enum-definition';
import { AppService } from '../app.service';

@WebSocketGateway({
  cors: { origin: '*' },
})
export class WebSocketHandler
  implements OnGatewayInit, OnGatewayConnection, OnGatewayDisconnect
{
  private readonly logger = new Logger(WebSocketHandler.name);

  constructor(private readonly appService: AppService) {}

  @WebSocketServer()
  server: Server;

  afterInit(server: any) {
    this.logger.log(`Server Started: ${server}`);
  }

  handleConnection(client: Socket) {
    this.logger.log(`Client connected: ${client.id}`);
  }

  handleDisconnect(client: Socket) {
    this.logger.log(`Client disconnected: ${client.id}`);
  }

  sendToClient(event: WebsocketEventType, data: object) {
    this.server.emit(event, data);
  }

  @SubscribeMessage(WebsocketEvent.MONAD_CHUNK_EVENT)
  async handleUDP(@MessageBody() data: any) {
    const doc = await this.appService.createFromUDP(data);
    this.sendToClient(WebsocketEvent.CLIENT_EVENT, doc);
  }

  @SubscribeMessage(WebsocketEvent.BPF_TRACE)
  async handleBPFTrace(@MessageBody() data: any) {
    const doc = await this.appService.saveBpfTrace(data);
    this.sendToClient(WebsocketEvent.BPF_TRACE, doc);
  }

  @SubscribeMessage(WebsocketEvent.SYSTEM_LOG)
  async handleSystemLog(@MessageBody() data: any) {
    const doc = await this.appService.saveSystemdLog(data);
    if (!doc) {
      return;
    }
    this.sendToClient(WebsocketEvent.SYSTEM_LOG, doc);
  }

  @SubscribeMessage(WebsocketEvent.OFF_CPU)
  async handleOffCpu(@MessageBody() data: any) {
    const doc = await this.appService.saveOffCpuEvent(data);
    this.sendToClient(WebsocketEvent.OFF_CPU, doc);
  }

  @SubscribeMessage(WebsocketEvent.SCHEDULER)
  async handleScheduler(@MessageBody() data: any) {
    const doc = await this.appService.saveSchedulerEvent(data);
    this.sendToClient(WebsocketEvent.SCHEDULER, doc);
  }

  @SubscribeMessage(WebsocketEvent.PERF_STAT)
  async handlePerfStatus(@MessageBody() data: any) {
    const doc = await this.appService.savePerfStatEvent(data);
    this.sendToClient(WebsocketEvent.PERF_STAT, doc);
  }

  @SubscribeMessage(WebsocketEvent.TURBO_STAT)
  async handleTurboStatus(@MessageBody() data: any) {
    const doc = await this.appService.saveTurboStatEvent(data);
    this.sendToClient(WebsocketEvent.TURBO_STAT, doc);
  }
}
