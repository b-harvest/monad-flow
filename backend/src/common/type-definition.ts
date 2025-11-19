import { NetworkEvent, WebsocketEvent } from './enum-definition';

export type WebsocketEventType =
  (typeof WebsocketEvent)[keyof typeof WebsocketEvent];

export type NetworkEventType = (typeof NetworkEvent)[keyof typeof NetworkEvent];
