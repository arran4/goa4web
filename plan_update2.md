To use the bus, we can add `BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error)` to `PrefetchExternalLinkTask`.
Wait, the `BackgroundTask` method runs asynchronously in the `backgroundtaskworker`.
Inside `BackgroundTask`, we need the `rawURL`. But `PrefetchExternalLinkTask` is currently just a stateless struct initialized once: `var prefetchExternalLinkTask = &PrefetchExternalLinkTask{TaskString: "externallink:prefetch"}`.
If we set `cd.Event().Task = t`, then `t` must contain the data (e.g. `rawURL`).
But if `prefetchExternalLinkTask` is a global variable, setting its fields is not thread-safe.

Ah! When we assign a task to `cd.Event().Task`, we can instantiate a *new* instance of a task struct that implements `BackgroundTasker`.
Wait, the `TaskHandler` automatically sets `cd.Event().Task = t` where `t` is the original task `prefetchExternalLinkTask` passed to `TaskHandler`.
Let's see `TaskHandler` again:
`v.SetEventTask(t)`
But `t` is `prefetchExternalLinkTask`.

If we want to pass data to the background task, how does the codebase usually do it? Let's search for `BackgroundTasker` implementations.
