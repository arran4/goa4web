package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
)

// DLQ stores messages in the database.
type DLQ struct{ Queries db.Querier }

// Record inserts the message into the dead letter table.
func (d DLQ) Record(ctx context.Context, message string) error {
	if d.Queries == nil {
		return fmt.Errorf("no db")
	}
	return d.Queries.SystemInsertDeadLetter(ctx, message)
}

// Get retrieves a message by ID.
func (d DLQ) Get(ctx context.Context, idStr string) (string, error) {
	if d.Queries == nil {
		return "", fmt.Errorf("no db")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return "", err
	}
	letter, err := d.Queries.SystemGetDeadLetter(ctx, int32(id))
	if err != nil {
		return "", err
	}
	return letter.Message, nil
}

// Delete removes a message by ID.
func (d DLQ) Delete(ctx context.Context, idStr string) error {
	if d.Queries == nil {
		return fmt.Errorf("no db")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}
	return d.Queries.SystemDeleteDeadLetter(ctx, int32(id))
}

// Register registers the database provider.
func Register(r *dlq.Registry) {
	r.RegisterProvider("db", func(_ *config.RuntimeConfig, q db.Querier) dlq.DLQ {
		return DLQ{Queries: q}
	})
}
