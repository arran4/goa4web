package file

import (
	"context"
	"fmt"
	"os"
	"sync"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/runtimeconfig"
)

// DLQ appends messages to a file.
type DLQ struct {
	Path string
	mu   sync.Mutex
}

// Record writes the message to the configured file.
func (f *DLQ) Record(_ context.Context, message string) error {
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

// Register registers the file provider.
func Register() {
	dlq.RegisterProvider("file", func(cfg runtimeconfig.RuntimeConfig, _ *dbpkg.Queries) dlq.DLQ {
		return &DLQ{Path: cfg.DLQFile}
	})
}
