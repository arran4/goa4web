The `workers/externallinkworker/worker.go` uses `cd := common.NewCoreData(ctx, q, cfg)` to get the `DownloadAndCacheImage` function.

So if I implement `BackgroundTasker` for `PrefetchExternalLinkTask`, I can do:

```go
type PrefetchExternalLinkTask struct {
    tasks.TaskString
    Config *config.RuntimeConfig
    URL    string
}

func (t *PrefetchExternalLinkTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
    // move the goroutine contents here
    // use common.NewCoreData(ctx, q, t.Config)
}
```

Wait, `ReloadExternalLinkTask` currently uses a `go func()`. Should I reply to the PR comment saying:
"It is possible to use `BackgroundTasker` and assign it to the event bus. However, both `PrefetchExternalLinkTask` and `ReloadExternalLinkTask` currently use an ad-hoc goroutine because the background task needs access to `common.CoreData` for caching images, which is easiest to obtain within the HTTP handler closure. If we refactored it to the bus, we would need to pass the configuration into the task struct and instantiate a new `*common.CoreData` inside the worker (like `externallinkworker` does). Let me know if you would like me to refactor both tasks to use the `BackgroundTasker` interface via the event bus!"

Actually, I can just reply directly to the comment and explain this.
