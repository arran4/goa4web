# workers

## Purpose

The `workers` directory contains asynchronous background processors. A 'worker' in Goa4Web is a component that listens to the central eventbus or a queue and executes tasks outside the critical path of an HTTP request.

**Why it exists:** To keep the web application fast, operations like sending emails, recounting forum posts, or auditing logs are offloaded to workers.
**What should become a worker:** Any long-running task, external API call, or batched database update that can be processed eventually and asynchronously without user intervention.
**Requirements:** A worker must handle context cancellation gracefully, connect to the eventbus, process its specific event type, and report errors without crashing the main application server.

## Structure and Components

This package is typically composed of core implementations, model definitions, and occasional testing utilities related specifically to this domain.

## Usage

Workers are initialized in `cmd/goa4web/main.go` and run as background goroutines. To dispatch work to them, you publish strongly typed events to the central `eventbus`.

```go
eventbus.Publish(ctx, "my_queue", myDataStruct)
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.
