import { Injectable, Logger } from '@nestjs/common';
import { HttpService } from '@nestjs/axios';
import { firstValueFrom } from 'rxjs';
import * as crypto from 'crypto';

@Injectable()
export class WebhooksService {
  private readonly logger = new Logger(WebhooksService.name);
  private webhookUrl = process.env.WEBHOOK_URL || 'http://go-service:8080/webhook';
  private webhookSecret = process.env.WEBHOOK_SECRET || 'my_webhook_secret';

  constructor(private httpService: HttpService) {}

  async sendEventParticipation(userId: number, eventId: number, action: 'join' | 'leave') {
    const payload = JSON.stringify({ userId, eventId, action, timestamp: new Date().toISOString() });
    
    // Create HMAC signature
    const signature = crypto.createHmac('sha256', this.webhookSecret).update(payload).digest('hex');

    try {
      await firstValueFrom(
        this.httpService.post(this.webhookUrl, payload, {
          headers: {
            'Content-Type': 'application/json',
            'X-Signature': signature,
          },
        })
      );
      this.logger.log(`Webhook sent for user ${userId} ${action} event ${eventId}`);
    } catch (error: any) {
      this.logger.error(`Failed to send webhook: ${error.message}`);
    }
  }
}
