import { Resolver, Mutation, Args, Int, Context } from '@nestjs/graphql';
import { PrismaService } from '../prisma/prisma.service';
import { EventParticipantModel } from './models/event.model';
import { UseGuards, UnauthorizedException } from '@nestjs/common';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { WebhooksService } from '../webhooks/webhooks.service';
import { EventLoader } from './loaders/event.loader';

@Resolver(() => EventParticipantModel)
export class EventResolver {
  constructor(
    private prisma: PrismaService,
    private webhooksService: WebhooksService,
    private eventLoader: EventLoader,
  ) {}

  @Mutation(() => EventParticipantModel)
  @UseGuards(JwtAuthGuard)
  async joinEvent(
    @Args('eventId', { type: () => Int }) eventId: number,
    @Context() context: any,
  ) {
    const userId = context.req.user?.userId;
    if (!userId) throw new UnauthorizedException();

    const participant = await this.prisma.eventParticipant.create({
      data: { eventId, userId },
    });

    // Notify Webhook Asynchronously
    this.webhooksService.sendEventParticipation(userId, eventId, 'join');

    return participant;
  }

  @Mutation(() => Boolean)
  @UseGuards(JwtAuthGuard)
  async leaveEvent(
    @Args('eventId', { type: () => Int }) eventId: number,
    @Context() context: any,
  ) {
    const userId = context.req.user?.userId;
    if (!userId) throw new UnauthorizedException();

    await this.prisma.eventParticipant.delete({
      where: { eventId_userId: { eventId, userId } },
    });

    // Notify Webhook Asynchronously
    this.webhooksService.sendEventParticipation(userId, eventId, 'leave');

    return true;
  }
}
