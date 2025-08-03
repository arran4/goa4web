# Email Queue Processing

Goa4Web stores outbound messages in a database table so they can be delivered by
an asynchronous worker. This decouples generating notifications from actually
sending mail.

## Enqueuing Messages

`internal/notifications.Notifier` renders email templates and enqueues the
result with `InsertPendingEmail`:

```go
_ = queries.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToUserID: sql.NullInt32{Int32: uid, Valid: true}, Body: string(msg), DirectEmail: false})
```

The admin interface uses the same method when previewing templates. Rows in the
`pending_emails` table contain the target user ID, message body and an
`error_count` column added in migration 0014.

## Background Worker

`EmailQueueWorker` runs in the `workers/emailqueue` package. It listens for
`eventbus.EmailQueueEvent` messages and processes queued emails whenever
signalled. After sending a message the worker waits at least
`EmailWorkerInterval` seconds before attempting the next delivery. The
`ProcessPendingEmail` function fetches one queued message with
`SystemListPendingEmails(ctx, db.SystemListPendingEmailsParams{Limit: 1, Offset: 0})`, loads the recipient address and sends the email via
the configured provider. Successful deliveries mark the row as sent. Failures
increment `error_count`. Once the count exceeds four the message is copied to the
DLQ (if configured) and removed from the queue.

## CLI Interaction

The `goa4web email queue` commands provide manual control:

- `list` prints unsent messages with their IDs and subjects
- `resend -id N` sends the specified message synchronously and marks it sent
- `delete -id N` removes a message without delivering it

`resend` loads the provider from the current configuration and calls
`Provider.Send` immediately instead of waiting for the worker.

## Configuration

`EMAIL_WORKER_INTERVAL` (`--email-worker-interval`) sets the minimum time in
seconds between sending queued emails. The default is `60`. Sending can be disabled entirely with
`EMAIL_ENABLED=false`. The worker currently processes a single email per run and
this batch size is not configurable.
