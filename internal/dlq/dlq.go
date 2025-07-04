package dlq

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

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

// FileDLQ appends messages to a file.
type FileDLQ struct {
	Path string
	mu   sync.Mutex
}

func (f *FileDLQ) Record(_ context.Context, message string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Path == "" {
		f.Path = "dlq.log"
	}
	fh, err := os.OpenFile(f.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fmt.Fprintln(fh, message)
	return err
}

// DBDLQ stores messages in the database.
type DBDLQ struct{ Queries *dbpkg.Queries }

func (d DBDLQ) Record(ctx context.Context, message string) error {
	if d.Queries == nil {
		return fmt.Errorf("no db")
	}
	return d.Queries.InsertWorkerError(ctx, message)
}
