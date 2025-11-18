import {
  WebSocketGateway,
  WebSocketServer,
  OnGatewayConnection,
  OnGatewayDisconnect,
  SubscribeMessage,
  MessageBody,
  ConnectedSocket,
  OnGatewayInit,
} from '@nestjs/websockets';
import { Server, Socket } from 'socket.io';
import { Logger } from '@nestjs/common';
import { WebsocketEventType } from '../common/type-definition';
import { WebsocketEvent } from '../common/enum-definition';
import { AppService } from '../app.service';

@WebSocketGateway({
  path: '/api/socket',
  transports: ['websocket'],
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
    const userId = client.handshake.headers['UserId'] as string;
    this.logger.log(
      `Client disconnected: ${client.id} (User: ${userId ?? 'unknown'})`,
    );
  }

  sendToClient(event: WebsocketEventType, data: object) {
    this.server.emit(event, data);
  }

  @SubscribeMessage(WebsocketEvent.CLIENT_EVENT)
  async handleClient(
    @MessageBody() data: any,
    @ConnectedSocket() client: Socket,
  ) {
    await client.join(WebsocketEvent.CLIENT_EVENT);
  }

  @SubscribeMessage(WebsocketEvent.TCP_EVENT)
  handleTCP(@MessageBody() data: any) {}

  @SubscribeMessage(WebsocketEvent.UDP_EVENT)
  async handleUDP(@MessageBody() data: any) {
    this.logger.log(`UDP Event Received: ${JSON.stringify(data)}`);
    await this.appService.createFromUDP(data);
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
