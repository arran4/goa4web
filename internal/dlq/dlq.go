package dlq

import (
	"context"
	"log"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
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
func RegisterLogDLQ() {
	RegisterProvider("log", func(runtimeconfig.RuntimeConfig, *dbpkg.Queries) DLQ {
		return LogDLQ{}
	})
}

func init() { RegisterLogDLQ() }
