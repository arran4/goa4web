package tasks

import (
	"context"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// BackgroundTasker is implemented by tasks that perform additional work
// after the HTTP response has been sent. The method should execute the
// background action and return a task published once the work is done.
// Returning nil indicates no follow-up event.
type BackgroundTasker interface {
	BackgroundTask(ctx context.Context, q *dbpkg.Queries) (Task, error)
}

// PostResultAction is kept for backward compatibility.
type PostResultAction = BackgroundTasker
