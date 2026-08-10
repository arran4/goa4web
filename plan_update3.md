So `PrefetchExternalLinkTask` needs to hold the `rawURL` state. But `TaskHandler(prefetchExternalLinkTask)` uses a predefined, static task. We can't mutate `prefetchExternalLinkTask.rawURL` in `Action()` because it would be a race condition.

Wait! The user can create a new task struct and assign it to `cd.Event().Task`.
```go
func (PrefetchExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	rawURL := r.FormValue("url")
	if rawURL == "" {
		return handlers.TextByteWriter("missing url")
	}

	cd.Event().Task = &PrefetchExternalLinkBackgroundTask{
		TaskString: "externallink:prefetch_background",
		URL:        rawURL,
	}

	return handlers.TextByteWriter("ok")
}

type PrefetchExternalLinkBackgroundTask struct {
	tasks.TaskString
	URL string
}

func (t *PrefetchExternalLinkBackgroundTask) BackgroundTask(ctx context.Context, q db.Querier) (tasks.Task, error) {
	// Need HTTPClient and CoreData things...
    // wait, we can't easily get `cd` in BackgroundTask. We just get `ctx` and `db.Querier`.
    // We would need to pass the HTTPClient or construct one.
    // In Goa4web, BackgroundTask runs entirely detached.
