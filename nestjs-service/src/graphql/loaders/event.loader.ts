import { Injectable, Scope } from '@nestjs/common';
import * as DataLoader from 'dataloader';
import { PrismaService } from '../../prisma/prisma.service';
import { EventParticipantModel } from '../models/event.model';

@Injectable({ scope: Scope.REQUEST })
export class EventLoader {
  constructor(private prisma: PrismaService) {}

  public readonly participantsLoader = new DataLoader<number, EventParticipantModel[]>(
    async (eventIds: readonly number[]) => {
      const participants = await this.prisma.eventParticipant.findMany({
        where: { eventId: { in: [...eventIds] } },
      });

      const groupedByEventId = eventIds.map((id) =>
        participants.filter((p) => p.eventId === id)
      );
      return groupedByEventId;
    }
  );
}
