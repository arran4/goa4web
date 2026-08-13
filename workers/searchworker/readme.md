# workers/searchworker

## Purpose

Package `searchworker` implements a specific background worker (`searchworker`). Workers are detached, asynchronous processors that respond to eventbus notifications, manage scheduled tasks, or process queues (like email or external link scanning). They handle heavy, long-running, or non-blocking tasks that should not delay the HTTP request-response cycle.

## Why It Exists

To keep the web application fast. Operations like sending emails, recounting forum posts, or auditing logs take time. Doing them during an HTTP request blocks the user from seeing their page load.

## What It Allows

It allows the system to fire-and-forget tasks. The web handler returns instantly, and the worker processes the heavy lifting in the background reliably.

## Structure and Components

The primary files and their general responsibilities include:

- `insert.go`
- `postupdate.go`
- `tokenize.go`
- `tokenize_test.go`
- `worker.go`
- `http.go`

### Exported Types and Interfaces

- **`InsertFunc`**:
- **`IndexedTask`** (Interface): Defines a core contract for this module.
- **`WordCount`**:
- **`IndexEventData`**:

### Exported Functions

- `InsertWords`
- `IsAlphanumericOrPunctuation`
- `BreakupTextToWords`
- `SearchWordIdsFromText`
- `InsertWordsToLinkerSearch`
- `InsertWordsToImageSearch`
- `InsertWordsToWritingSearch`
- `InsertWordsToForumSearch`
- `Worker`

## Usage Examples

Workers subscribe to topics on the `eventbus`. To trigger a worker, a handler publishes an event to the bus. The worker receives the payload, executes its logic, and optionally publishes a new event (e.g. via Websockets) when complete.

```go
import "github.com/arran4/goa4web/internal/eventbus"

// Trigger a background task from a handler
eventbus.Publish(ctx, "my_queue_topic", myDataStruct)
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.
