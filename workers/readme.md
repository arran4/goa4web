# workers

## Purpose

The `workers` directory contains asynchronous background processors. A 'worker' in Goa4Web is a component that listens to the central eventbus or a queue and executes tasks outside the critical path of an HTTP request.

## Why It Exists

To keep the web application fast. Operations like sending emails, recounting forum posts, or auditing logs take time. Doing them during an HTTP request blocks the user from seeing their page load.

## What It Allows

It allows the system to fire-and-forget best-effort tasks. The web handler can
return before the worker processes the heavy lifting. Event-bus work remains
volatile and may be lost on a crash or forced shutdown.

## Structure and Components

The primary files and their general responsibilities include:

- `workers.go`

### Exported Functions

- `Start`, which returns a managed `Group` used for cancellation and bounded waits.

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
- **Shutdown**: Accepted in-memory work is allowed to drain within the server's
  shutdown deadline. At the deadline workers are cancelled and resource closure
  continues; work that ignored cancellation is not awaited indefinitely.
