import { Body, Controller, Get, HttpStatus, Post, Res } from '@nestjs/common';
import type { Response } from 'express';
import { AppService } from './app.service';
import { NetworkEventType } from './common/type-definition';

@Controller('/api')
export class AppController {
  constructor(private readonly appService: AppService) {}

  @Get()
  async getAll(@Res() response: Response) {
    await this.appService.getAll();
    response.status(HttpStatus.OK).send();
  }

  @Post('/outbound-message')
  async createOutboundMessage(
    @Body() body: { type: NetworkEventType; data: any },
    @Res() response: Response,
  ) {
    const result = await this.appService.createFromUDP(body);
    response.status(HttpStatus.CREATED).send(result);
  }
}
