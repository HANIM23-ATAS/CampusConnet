import { Module } from '@nestjs/common';
import { PrismaModule } from './prisma/prisma.module';
import { UsersModule } from './users/users.module';
import { AuthModule } from './auth/auth.module';
import { EventsModule } from './events/events.module';
import { GraphQLModule } from '@nestjs/graphql';
import { ApolloDriver, ApolloDriverConfig } from '@nestjs/apollo';
import { EventGraphqlModule } from './graphql/event-graphql.module';
import { WebhooksModule } from './webhooks/webhooks.module';

@Module({
  imports: [
    GraphQLModule.forRoot<ApolloDriverConfig>({
      driver: ApolloDriver,
      autoSchemaFile: true,
      context: ({ req }) => ({ req }),
    }),
    PrismaModule,
    UsersModule,
    AuthModule,
    EventsModule,
    EventGraphqlModule,
    WebhooksModule,
  ],
})
export class AppModule {}
