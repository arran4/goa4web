# internal/eventbus

## Purpose

Package `eventbus` provides internal, non-exported utilities and service integrations specific to `eventbus`.

## Why It Exists

To encapsulate the logic necessary for this specific operational domain, ensuring modularity within the codebase.

## What It Allows

It allows the system to remain decoupled. Code outside this package can rely on its exported API without worrying about its internal implementation details.

## Structure and Components

The primary files and their general responsibilities include:

- `benchmark_test.go`
- `eventbus.go`
- `eventbus_test.go`

### Exported Types and Interfaces

- **`TaskEvent`**:
  - Methods: `Type`
- **`EmailQueueEvent`**:
  - Methods: `Type`
- **`DigestRunEvent`**:
  - Methods: `Type`
- **`Bus`**:
  - Methods: `Subscribe`, `Publish`, `Shutdown`
- **`MessageType`**:
- **`Message`** (Interface): Defines a core contract for this module.
- **`Envelope`**:
  - Methods: `Ack`

### Exported Functions

- `BenchmarkShutdown`
- `NewBus`
- `TestBus_Shutdown`
- `TestBus_Ack`
- `TestBus_Backpressure`
- `TestBus_ShutdownContext`
- `TestNewBus`
- `TestSubscribe`
- `TestPublish`
- `TestPublish_NonBlocking`
- `TestShutdown`
- `TestShutdown_Timeout`
- `TestSyncPublish`
- `TestConcurrentAccess`

## Usage Examples

The eventbus is the central nervous system for async work. Use it to decouple HTTP handlers from slow background tasks.

```go
import "goa4web/internal/eventbus"

// Publisher
eventbus.Publish(ctx, "UserRegistered", user.ID)

// Subscriber (typically inside a worker init)
eventbus.Subscribe("UserRegistered", func(ctx context.Context, payload interface{}) {
    userID := payload.(int)
    // process it
})
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
