# Dead Letter Queue

This document describes the dead letter queue (DLQ) mechanism used for failed background tasks.

## Interface

The DLQ system is defined by the `dlq.DLQ` interface:

```
type DLQ interface {
        Record(ctx context.Context, message string) error
}
```

Providers implement `Record` to persist messages for later inspection.

## Default Providers

The `internal/dlq/dlqdefaults` package registers the stable implementations:

- **log** – writes messages to the application log.
- **file** – appends messages to a single log file specified by `DLQ_FILE`.
- **dir** – stores each message as a file under the directory given by `DLQ_FILE`.
- **db** – inserts messages into the database using the `dead_letters` table.
- **email** – emails messages to administrator addresses using the configured mail provider.

Multiple providers can be combined by setting `DLQ_PROVIDER` to a comma‑separated list.

## Configuration Keys

| Key | Description |
|-----|-------------|
| `DLQ_PROVIDER` | Selects one or more providers (`log`, `file`, `dir`, `db`, `email`). If empty the log provider is used. |
| `DLQ_FILE` | Path used by the `file` and `dir` providers. Defaults to `dlq.log` or `dlq/` when not set. |

## Worker Integration

Workers forward failed operations to the DLQ:

- `EmailQueueWorker` records an entry when sending a queued email fails after several attempts.
- `Notifier.BusWorker` calls `dlqRecordAndNotify` to store failures delivering notifications.

These hooks ensure problems are captured for follow‑up without disrupting normal processing.
