# workers

## Purpose

The `workers` directory contains asynchronous background processors. A 'worker' in Goa4Web is a component that listens to the central eventbus or a queue and executes tasks outside the critical path of an HTTP request.

## Context and Use Cases (How and Why)

**Why it exists:** To keep the web application fast. Operations like sending emails, recounting forum posts, or auditing logs take time. Doing them during an HTTP request blocks the user.
**What this allows:** It allows the system to fire-and-forget tasks. The web handler returns instantly, and the worker processes the heavy lifting in the background.
**How to use it:** Workers subscribe to topics on the `eventbus`. To trigger a worker, a handler publishes an event to the bus. The worker receives the payload, executes its logic, and optionally publishes a new event (e.g. via Websockets) when complete.

## Structure and Components

The primary files and their general responsibilities include:

- `workers.go`

### Exported Functions

- `Start`

## Usage Examples

Workers are initialized in `cmd/goa4web/main.go` and run as background goroutines. To dispatch work to them, you publish strongly typed events to the central `eventbus`.

```go
import "goa4web/internal/eventbus"

// Trigger a background task
eventbus.Publish(ctx, "my_queue_topic", myDataStruct)
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.
