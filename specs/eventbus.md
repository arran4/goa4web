# Event Bus Overview

This document describes how the internal event bus works and how events
are consumed by the notification worker.

## Event Structure

Events are defined in `internal/eventbus/eventbus.go`. The core message interface is:

```go
type Message interface {
	Type() MessageType
}
```

Two main message types are supported:
- `TaskMessageType` for application actions (`TaskEvent`)
- `EmailQueueMessageType` for email processing triggers (`EmailQueueEvent`)

### TaskEvent

`TaskEvent` contains contextual information about a user action:

- `Path` – a path uniquely identifying the affected object; often the URL path but not always.
- `Task` – the `tasks.Task` instance associated with the action.
- `UserID` – identifier of the user performing the action.
- `Time` – timestamp when the event occurred.
- `Data` – optional key/value map used when rendering templates.
- `Outcome` – status string (e.g. "success").

## Subscription Model

Subscriptions are stored in the `subscriptions` table and are matched
using path patterns. A subscription pattern is built from the task name
and the request path. For example `reply:/blog/a/b` and the wildcard
variants returned by `buildPatterns`. Subscribers may register for
`email` or `internal` notifications.

`collectSubscribers` queries the database for every matching pattern and
returns the union of user IDs for the chosen delivery method. Events can
also specify a `target` item through `Event.Data` which is used to link
notifications to a specific record.

## Bus Implementation

The `Bus` struct manages subscribers and dispatching.

- `Subscribe(types ...MessageType)` returns a channel that receives matching messages.
- `Publish(msg Message)` sends a message to all matching subscribers non-blocking (dropping messages if channel is full).

The bus and its acknowledgements are process-local and volatile. `Envelope.Ack`
only releases in-memory shutdown accounting; it does not persist the message or
prove that work can resume after a process exit. Reliable subscriptions apply
backpressure while the process is running, but are not durable queues.

## Remote image cache recovery

Embedded and standalone workers reconcile due remote image cache entries into the
durable job queue at startup and periodically. Both reconciliation and the
immediate enqueue use the cache entry ID and leave existing job retry schedules
intact. Workers also accept previously queued URL-keyed image jobs. Cancelling
a queued image updates its cache entry in the same transaction, so recovery
respects cancellation even after the job history is pruned. The worker writes
the original and thumbnail to content-addressed keys outside the transaction.
The active lease gates the transaction that selects which keys a signed cache
URL serves. A late write by a replaced worker cannot select its bytes. Existing
cache rows with unqualified checksums still resolve their original filenames.
Fetch diagnostics identify the operation or upstream status without exposing
the remote image URL.
Remote responses have a 50 MiB read limit, including when the server omits
the length. Raster decoding is limited to 40 million pixels and 16,384
pixels per side.

## Observed external links

The embedded server treats successful task events from older producers as
inputs to the durable external-link queue. Forum opening comments, replies,
appends and edits record eligible link jobs in the same transaction as their
content. Blog, image board, linker, news and writing replies use the same
transaction boundary. Worker wake-ups follow the commit. Task events indicate
when their link jobs have already been recorded, so the compatibility worker does not
enqueue them again. Observing a link preserves the schedule of an active
retry and respects its cooldown.

The worker fetches outside the job transaction and checks the current lease
when publishing metadata or its image reference. It rejects unsuccessful HTTP
responses before parsing metadata. HTTP 429 and 503 can extend the retry delay
using `Retry-After`, capped at five minutes. Exhausted transient failures use
the configured cooldown. Invalid destinations and permanent HTTP failures
remain suppressed until an explicit retry.

## Durable job outcomes

Completion, retry and failure require an unexpired lease. An expired
attempt leaves the queue entry available for another worker to claim.

Terminal failure, cooldown and administrator inbox updates commit together.
Each recipient has one alert for a work type, target and error class. A
repeated failure updates the alert count and latest time. An acknowledged
alert reopens after the debounce window. The grouping key and summary omit
the target URL. Pruning terminal job history leaves the alert intact.

## Shutdown

Calling `Bus.Shutdown(ctx)` stops new publications, closes delivery to each
subscriber, and waits for accepted envelopes to be acknowledged. Consumers can
process buffered envelopes while their channels drain. A subscriber that closes
early acknowledges its remaining buffered envelopes without processing them.
The call returns when accounting reaches zero or the context is cancelled.

## BusWorker

`notifications.BusWorker` subscribes to an `eventbus.Bus` (specifically for `TaskMessageType`) and processes
each event in sequence. It invokes `processEvent` which renders email and
internal notifications based on interfaces implemented by the event's
task.

Subscriber lists are built by `notifySubscribers`. The worker calls
`buildPatterns` with the task and path to generate all matching
subscription patterns. It then calls `collectSubscribers` for `email` and
`internal` methods to produce unique recipient sets before sending the
notifications.
