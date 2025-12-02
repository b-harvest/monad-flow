import {
  Body,
  Controller,
  Get,
  HttpStatus,
  Param,
  Post,
  Query,
  Res,
} from '@nestjs/common';
import type { Response } from 'express';
import { AppService } from './app.service';
import { NetworkEventType } from './common/type-definition';
import { WebSocketHandler } from './websocket/websocket.handler';
import { WebsocketEvent } from './common/enum-definition';
import { GetLogsDto } from './dto/get-logs.dto';

@Controller('/api')
export class AppController {
  constructor(
    private readonly appService: AppService,
    private readonly websocketHandler: WebSocketHandler,
  ) {}

  @Get('/app-message/:id')
  async getAppMessage(@Param('id') id: string, @Res() response: Response) {
    const result = await this.appService.getAppMessage(id);
    response.status(HttpStatus.OK).send(result);
  }

  @Get('/leader/:round')
  async getLeader(
    @Param('round') round: number,
    @Query('range') range: number,
    @Res() response: Response,
  ) {
    const result = await this.appService.getLeaders(round, range);
    response.status(HttpStatus.OK).send(result);
  }

  @Get('/logs/:type')
  async getLogsByTimeRange(
    @Param('type') type: string,
    @Query() query: GetLogsDto,
    @Res() response: Response,
  ) {
    const startTime = new Date(query.from);
    const endTime = new Date(query.to);
    const result = await this.appService.getLogsByTimeRange(
      startTime,
      endTime,
      type,
    );
    response.status(HttpStatus.OK).send(result);
  }

  @Post('/outbound-message')
  async createOutboundMessage(
    @Body()
    body: {
      type: NetworkEventType;
      data: any;
      timestamp: number;
      appMessageHash?: string;
    },
    @Res() response: Response,
  ) {
    const createdDoc = await this.appService.createFromUDP(body);
    const result = createdDoc.toObject ? createdDoc.toObject() : createdDoc;

    let websocketPayload = result;
    if (result.messageType === 1) {
      const { data, ...summary } = result;
      websocketPayload = summary;
    }
    this.websocketHandler.sendToClient(
      WebsocketEvent.OUTBOUND_ROUTER,
      websocketPayload,
    );

    response.status(HttpStatus.CREATED).send();
  }

  @Post('/leader')
  async createLeader(
    @Body()
    body: {
      epoch: number;
      round: number;
      node_id: string;
      cert_pubkey: string;
      stake: string;
      timestamp: number;
    },
    @Res() response: Response,
  ) {
    await this.appService.saveLeader(body);
    response.status(HttpStatus.CREATED).send();
  }
}
