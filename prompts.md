# Architectural Refactoring for Background Tasks

## Current Situation
In the `goa4web` application, background tasks are generally offloaded using the `tasks.BackgroundTasker` interface, which is picked up by the `backgroundtaskworker`.

However, some background operations (like fetching OpenGraph metadata in `ReloadExternalLinkTask` and `PrefetchExternalLinkTask`) currently rely on ad-hoc `go func()` goroutines.

### Why ad-hoc goroutines were used:
1. **Context and Dependencies:** The `tasks.BackgroundTasker` interface currently provides only a database querier (`db.Querier`) and a `context.Context` to the worker method:
   `BackgroundTask(ctx context.Context, q db.Querier) (Task, error)`
2. **Missing Application State:** `externallink` tasks require the application's configuration (`config.RuntimeConfig`) to initialize the image caching provider (`cd.DownloadAndCacheImage`), and a configured HTTP client to fetch external URLs. These are easily available inside the HTTP handler (via `*common.CoreData`) but are missing from the `BackgroundTasker` interface.

## Proposed Architectural Solution

To achieve consistency and properly use the event bus for these tasks, the stack should be refactored to supply broader application context to background workers.

### 1. Enhancing the Background Task Context
The `backgroundtaskworker` is responsible for executing tasks. Instead of just passing `db.Querier`, we should pass a comprehensive struct (e.g., `TaskContext` or a background-safe `*common.CoreData`).

Since `common.NewCoreData(ctx, q, cfg)` can instantiate a fresh, request-independent CoreData object (as seen in `workers/externallinkworker/worker.go`), we can pass the `config.RuntimeConfig` to the `backgroundtaskworker`.

**Proposed changes to `backgroundtaskworker`:**
- Pass the `*config.RuntimeConfig` into the `backgroundtaskworker` during initialization.
- Modify the `tasks.BackgroundTasker` interface to accept `*common.CoreData` or provide a way for the task to access the configuration.

**Alternative (Simpler) Approach:**
Tasks that implement `BackgroundTasker` can explicitly hold the required state.
```go
type PrefetchExternalLinkTask struct {
    tasks.TaskString
    URL    string
    Config *config.RuntimeConfig
}

func (t *PrefetchExternalLinkTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
    // Instantiate a background-safe CoreData instance
    cd := common.NewCoreData(ctx, q, t.Config)

    info, err := opengraph.Fetch(t.URL, cd.HTTPClient())
    // ... cache image and update DB ...
    return nil, nil
}
```
In the HTTP Handler:
```go
cd.Event().Task = &PrefetchExternalLinkTask{
    TaskString: "externallink:prefetch_background",
    URL:        rawURL,
    Config:     cd.Config, // Pass the global config down to the task
}
```

### 2. Identifying Other Areas for Refactoring
Ad-hoc `go func()` routines exist in a few other places that could benefit from this cleanup or at least a standardized background context.

1. **`handlers/externallink/tasks.go`**:
   - `ReloadExternalLinkTask`
   - `PrefetchExternalLinkTask`
   *Refactor:* Replace both with instances of a new struct implementing `BackgroundTasker` and assign them to `cd.Event().Task`.

2. **`handlers/admin/server_shutdown_task.go`**:
   - `TaskServerShutdown` uses an ad-hoc goroutine to delay shutdown. This is arguably acceptable since the server is tearing down and the event bus is stopping, but it's worth noting.

3. **`handlers/admin/adminImageCachePage.go`**:
   - Uses `go func()` to traverse and delete cached image files.
   *Refactor:* Could be offloaded to a true background worker (e.g., `ClearImageCacheTask` implementing `BackgroundTasker`) to prevent hanging HTTP responses or dangling goroutines.

4. **`handlers/imagebbs/imagebbsBoardPage.go`**:
   - Schedules background thumbnail generation directly in a goroutine if no event manager exists.

## Instructions for Implementation
If you ask the agent to implement this, use a prompt similar to:
> "Please refactor `ReloadExternalLinkTask` and `PrefetchExternalLinkTask` in `handlers/externallink/tasks.go` to use the event bus. Create new structs that implement `tasks.BackgroundTasker`. Have the HTTP `Action` handlers assign these structs to `cd.Event().Task` and pass along `cd.Config` and the URL. In the `BackgroundTask` method, instantiate a background `*common.CoreData` using `common.NewCoreData(ctx, q, t.Config)` to perform the OpenGraph fetch and image caching. Ensure the ad-hoc goroutines are removed. Additionally, once the OpenGraph data is successfully fetched and saved to the database, use the application's WebSocket (wss) infrastructure to push the updated metadata/card back to the connected client so the UI can reflect the prefetched content immediately."
