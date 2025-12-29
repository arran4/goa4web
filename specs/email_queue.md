# Email Queue

Outbound emails are queued in the `pending_emails` table and processed asynchronously.

## Flow

1. **Enqueue**: `InsertPendingEmail` adds the message body and recipient ID.
2. **Worker**: `EmailQueueWorker` (triggered by event bus or timer):
    - Fetches pending emails (ordered by priority/time).
    - Sends via the configured Email Provider.
    - On success: updates `sent_at`.
    - On failure: increments `error_count`.
    - Max retries (4) exceeded: moves to Dead Letter Queue (DLQ).

## Configuration

- `EMAIL_WORKER_INTERVAL`: Minimum delay between sends.
- `EMAIL_ENABLED`: Master switch.

## CLI

- `email queue list`: View pending.
- `email queue resend`: Force send immediately.
- `email queue delete`: Drop message.
