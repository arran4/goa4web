package dlq

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"
	"github.com/segmentio/ksuid"
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

// DirDLQ writes each message to a new file inside Dir using a KSUID filename.
type DirDLQ struct {
	Dir string
}

func (d *DirDLQ) Record(_ context.Context, message string) error {
	if d.Dir == "" {
		d.Dir = "dlq"
	}
	if err := os.MkdirAll(d.Dir, 0o755); err != nil {
		return err
	}
	name := ksuid.New().String() + ".txt"
	path := filepath.Join(d.Dir, name)
	return os.WriteFile(path, []byte(message+"\n"), 0o644)
}

// EmailDLQ sends DLQ messages to administrator emails using the given provider.
type EmailDLQ struct {
	Provider email.Provider
	Queries  *dbpkg.Queries
}

func (e EmailDLQ) Record(ctx context.Context, message string) error {
	if e.Provider == nil {
		return fmt.Errorf("no email provider")
	}
	for _, addr := range emailutil.GetAdminEmails(ctx, e.Queries) {
		if err := e.Provider.Send(ctx, addr, "DLQ message", message, ""); err != nil {
			log.Printf("dlq email: %v", err)
		}
	}
	return nil
}
