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
      `mongodb://${process.env.MONGO_ROOT_USERNAME}:${process.env.MONGO_ROOT_PASSWORD}@localhost:27017/${process.env.MONGO_DATABASE}?authSource=admin`,
    ),
  ],
  controllers: [AppController],
  providers: [WebSocketHandler, AppService],
})
export class AppModule {}
