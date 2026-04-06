#!/bin/sh
set -e

echo "Deploying Prisma migrations..."
npx prisma migrate deploy

echo "Starting NestJS server..."
npm run start:prod
