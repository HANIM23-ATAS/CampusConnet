import { ObjectType, Field, Int } from '@nestjs/graphql';

@ObjectType()
export class EventParticipantModel {
  @Field(() => Int)
  id!: number;

  @Field(() => Int)
  eventId!: number;

  @Field(() => Int)
  userId!: number;
}
