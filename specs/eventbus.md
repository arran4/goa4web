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

Message types include:
- `TaskMessageType` for application actions (`TaskEvent`)
- `EmailQueueMessageType` for email processing triggers (`EmailQueueEvent`)
- `NotificationsChangedMessageType` for recipient inbox invalidation (`NotificationsChanged`)

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

## Image board thumbnails

Image board uploads record thumbnail work in the same transaction as the post
and upload metadata. Embedded and standalone workers use the existing durable
queue. They regenerate the 200-pixel square thumbnail from the original stored
under its content hash, preserving existing thumbnail URLs. A replacement worker
can reclaim the job after its lease expires. Repeated processing writes the
same bytes to the same key.

## Observed external links

The embedded server treats successful task events from older producers as
inputs to the durable external-link queue. Forum opening comments, replies,
appends and edits record eligible link jobs in the same transaction as their
content. Blog, image board, linker, news and writing replies and edits use the
same transaction boundary. Signed `/goto` reload requests and authenticated
`/goto/prefetch` requests enqueue durable external-link work after validating
destinations against internal/private hosts instead of attaching volatile
background tasks when durable jobs are enabled. Worker wake-ups follow the
commit. Task events indicate when their link jobs have already been recorded, so
the compatibility worker does not enqueue them again. Observing a link preserves
the schedule of an active retry and respects its cooldown.

The worker fetches outside the job transaction and checks the current lease
when publishing metadata or its image reference. It rejects unsuccessful HTTP
responses before parsing metadata. HTTP 429 and 503 can extend the retry delay
using `Retry-After`, capped at five minutes. Exhausted transient failures use
the configured cooldown. Invalid destinations and permanent HTTP failures
remain suppressed until an explicit retry.

## Durable search indexing

Forum opening comments, replies, appends and edits record a search revision
and coalesce its job in the source transaction. Blog, image board, linker,
news and writing reply creation and edits use the same transaction boundary.
Writing creation and edits, admin comment edit/delete/deactivate/restore,
admin linker updates, link approvals, image board uploads/updates/deletions,
and CLI content deactivation/restoration commands commit search revisions with
their source changes. When enabled, `searchworker.Worker` converts any legacy
bus index events into durable search revisions rather than mutating search
tables without a lease. A search job reads the current source and replaces its
per-item terms. The term replacement and indexed revision commit together under
an active job lease. A superseded revision cannot advance the indexed revision.
Deleted or empty content clears its previous terms.

## Admin job controls and remaining bus workers

Operators can list, inspect (`goa4web jobs show` and `/admin/jobs?id=...`),
enqueue URLs, retry, cancel, and purge terminal jobs via CLI and authenticated
admin routes. Enqueueing an unused external URL validates the destination and
creates or reuses its `external_links` record atomically with the durable job.
Standalone (`goa4web worker`) and embedded durable loops share database leases
across restarts and forced crashes (`SIGKILL` or second `Ctrl+C`), reclaiming
expired claims safely.

Remaining non-durable bus workers keep their intended roles: `emailqueue` and
DLQ remain database-backed; `notifications` keeps best-effort fan-out and
scheduled digests; `postcountworker` recomputes regenerable thread/topic
statistics; `auditworker` and `logworker` record operational logs best-effort;
and `backgroundtaskworker` serves admin full-index remake tasks and fallback
behavior when no durable job service is configured.

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

## Live notification refresh

After the notifier stores an inbox entry, it can publish a best-effort signal
for the recipient. The authenticated `/ws/events` connection and the legacy
`/ws/notifications` route filter that signal by session user. The browser gets
only the notification event kind and timestamp.

The browser fetches server-rendered fragments from the page's
`data-live-url` endpoints on a signal and whenever the connection opens.
Those endpoints handle authorization and rendering. Clients can re-fetch
the current state independently of the delivery or ordering of socket events.

## Live forum refresh

The generic socket accepts a single visible thread scope and up to 12 block
link cards on that page. Forum replies publish a committed resource signal.
The gateway checks current account, topic, thread, and reply visibility before
delivering a thread invalidation. The receiving page rechecks access and
compares its displayed activity position with visible replies to render a
refresh prompt. The prompt leaves the reply form and read marker intact.

An external-link card is identified by its comment and the digest of its
canonical URL. The server resolves the key against the current visible
comment. The gateway observes stored metadata and active job state, so it
detects changes from both an embedded and a standalone worker. It rechecks
source visibility before sending a signal. Both consumers send only the kind
and time to the browser. On connection or reconnection, the browser reads the
authorized fragments to reconcile any changes it missed.
