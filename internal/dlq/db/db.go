package db

import (
	"context"
	"fmt"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
)

// DLQ stores messages in the database.
type DLQ struct{ Queries *dbpkg.Queries }

// Record inserts the message into the dead letter table.
func (d DLQ) Record(ctx context.Context, message string) error {
	if d.Queries == nil {
		return fmt.Errorf("no db")
	}
	return d.Queries.SystemInsertDeadLetter(ctx, message)
}

// Register registers the database provider.
func Register(r *dlq.Registry) {
	r.RegisterProvider("db", func(_ *config.RuntimeConfig, q *dbpkg.Queries) dlq.DLQ {
		return DLQ{Queries: q}
	})
}
