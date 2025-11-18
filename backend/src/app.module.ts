import { Module } from '@nestjs/common';
import { WebSocketHandler } from './websocket/websocket.handler';
import { MongooseModule } from '@nestjs/mongoose';
import { ModelModule } from './config/model.module';
import { AppService } from './app.service';
import { AppController } from './app.controller';

@Module({
  imports: [
    ModelModule,
    MongooseModule.forRoot(
      'mongodb://root:bharvest_password!@localhost:27017/bharvest?authSource=admin',
    ),
  ],
  controllers: [AppController],
  providers: [WebSocketHandler, AppService],
})
export class AppModule {}
