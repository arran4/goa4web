package dlq

import (
	"context"
	"log"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

// DLQ records failed asynchronous operations.
type DLQ interface {
	Record(ctx context.Context, message string) error
}

// LogDLQ writes messages to the application log.
type LogDLQ struct{}

func (LogDLQ) Record(_ context.Context, message string) error {
	log.Printf("dlq: %s", message)
	return nil
}

// RegisterLogDLQ registers the log provider.
func RegisterLogDLQ(r *Registry) {
	r.RegisterProvider("log", func(*config.RuntimeConfig, *dbpkg.Queries) DLQ {
		return LogDLQ{}
	})
}
