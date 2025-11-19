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

  @SubscribeMessage(WebsocketEvent.TCP_EVENT)
  handleTCP(@MessageBody() data: any) {}

  @SubscribeMessage(WebsocketEvent.UDP_EVENT)
  async handleUDP(@MessageBody() data: any) {
    const savedData = await this.appService.createFromUDP(data);
    this.sendToClient(WebsocketEvent.CLIENT_EVENT, savedData);
  }

  @SubscribeMessage(WebsocketEvent.JOURNAL_CTL_EVENT)
  handleJournalCtl(@MessageBody() data: any) {}

  @SubscribeMessage(WebsocketEvent.OFF_CPU_TIME_EVENT)
  handleOffCpuTime(@MessageBody() data: any) {}

  @SubscribeMessage(WebsocketEvent.PERF_SCHED_EVENT)
  handlePerfSched(@MessageBody() data: any) {}

  @SubscribeMessage(WebsocketEvent.PERF_STAT_EVENT)
  handlePerfStat(@MessageBody() data: any) {}

  @SubscribeMessage(WebsocketEvent.TURBO_STAT_EVENT)
  handleTurboStat(@MessageBody() data: any) {}
}
