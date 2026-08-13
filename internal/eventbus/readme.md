# internal/eventbus

## Purpose

Package `eventbus` provides internal, non-exported utilities and service integrations specific to `eventbus`.

## Structure and Components

The primary files and their general responsibilities include:

- `benchmark_test.go`
- `eventbus.go`
- `eventbus_test.go`

### Exported Types and Interfaces

- **`DigestRunEvent`**:
  - Methods: `Type`
- **`Bus`**:
  - Methods: `Subscribe`, `Publish`, `Shutdown`
- **`MessageType`**:
- **`Message`** (Interface): Defines a core contract for this module.
- **`Envelope`**:
  - Methods: `Ack`
- **`TaskEvent`**:
  - Methods: `Type`
- **`EmailQueueEvent`**:
  - Methods: `Type`

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

## Usage

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
