import { Controller, Get, HttpStatus, Inject, Res } from '@nestjs/common';
import type { Response } from 'express';
import { AppService } from './app.service';

@Controller('/api')
export class AppController {
  constructor(
    @Inject('AppService')
    private readonly appService: AppService,
  ) {}

  @Get()
  async getAll(@Res() response: Response) {
    await this.appService.getAll();
    response.status(HttpStatus.OK).send();
  }
}
