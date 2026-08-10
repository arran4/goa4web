The user wants to know if it's possible to add the prefetch task to the event bus rather than spawning a goroutine. Yes, it is. The `goa4web` application has a `backgroundtaskworker` that picks up tasks assigned to `cd.Event().Task`.
The memory state says: "In `goa4web`, background tasks are managed via `internal/eventbus`. To offload CPU-bound work from HTTP handlers, assign a struct implementing the `tasks.BackgroundTasker` interface to the request's event (`cd.Event().Task`), which will be picked up and executed asynchronously by the `backgroundtaskworker`."

Wait, the `ReloadExternalLinkTask` currently uses a `go func()`. Should I refactor both or just change `PrefetchExternalLinkTask`?
Let's see what `tasks.BackgroundTasker` looks like. I'll read `internal/tasks/tasks.go` or similar to see the interface.
