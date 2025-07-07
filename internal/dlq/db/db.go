package db

import (
	"context"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/runtimeconfig"
)

// DLQ stores messages in the database.
type DLQ struct{ Queries *dbpkg.Queries }

// Record inserts the message into the worker error table.
func (d DLQ) Record(ctx context.Context, message string) error {
	if d.Queries == nil {
		return fmt.Errorf("no db")
	}
	return d.Queries.InsertWorkerError(ctx, message)
}

// Register registers the database provider.
func Register() {
	dlq.RegisterProvider("db", func(_ runtimeconfig.RuntimeConfig, q *dbpkg.Queries) dlq.DLQ {
		return DLQ{Queries: q}
	})
}
