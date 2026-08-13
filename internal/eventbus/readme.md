# Internal event bus

## Why it exists

The event bus decouples request-time publishers from in-process background
consumers. It is appropriate for best-effort signals where the work may happen
after the HTTP response. It is not durable: use the database-backed task or queue
mechanisms when work must survive a restart.

## Message model

Every message implements `Message` by returning a `MessageType`. The built-in
messages are `TaskEvent`, `EmailQueueEvent`, and `DigestRunEvent`. A `Bus` is an
instance passed to publishers and workers; there are no package-level publish or
subscribe functions. Subscribers receive `Envelope` values and must call `Ack`
once processing is finished so `Shutdown` can drain outstanding delivery.

## Publishing and consuming

```go
bus := eventbus.NewBus()
events := bus.Subscribe(eventbus.TaskMessageType)

go func() {
    for envelope := range events {
        event, ok := envelope.Msg.(eventbus.TaskEvent)
        if ok {
            // Process event. Record/log failures according to the worker policy.
            _ = event
        }
        envelope.Ack()
    }
}()

err := bus.Publish(eventbus.TaskEvent{
    Path: "/forum/topic/42",
    Task: task,
    UserID: userID,
    Time: time.Now(),
})
```

Subscribe before publishing. Delivery is non-blocking and a full subscriber
buffer drops the message, so do not use this API for required or retryable work.
On shutdown, stop publishers and call `bus.Shutdown(ctx)` with a deadline.
