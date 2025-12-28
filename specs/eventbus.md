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

## Shutdown

Calling `Bus.Shutdown(ctx)` stops new publications and waits for all
queued messages on subscriber channels to drain. The call returns when
either all pending events are processed or the context is cancelled.

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
