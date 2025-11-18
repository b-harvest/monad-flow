import { UDPEvent, WebsocketEvent } from './enum-definition';

export type WebsocketEventType =
  (typeof WebsocketEvent)[keyof typeof WebsocketEvent];

export type UDPEventType = (typeof UDPEvent)[keyof typeof UDPEvent];
