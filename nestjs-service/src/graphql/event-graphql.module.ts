import { Module } from '@nestjs/common';
import { EventResolver } from './event.resolver';
import { WebhooksModule } from '../webhooks/webhooks.module';
import { EventLoader } from './loaders/event.loader';

@Module({
  imports: [WebhooksModule],
  providers: [EventResolver, EventLoader],
})
export class EventGraphqlModule {}
