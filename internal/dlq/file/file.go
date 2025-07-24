package file

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
)

// DLQ appends messages to a file.
type DLQ struct {
	Path string
	mu   sync.Mutex
}

// fileSeparator marks the boundary around each recorded message.
const fileSeparator = "-----"

var appendFile = func(name string, data []byte) error {
	fh, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fh.Write(data)
	return err
}

// Record writes the message to the configured file.
func (f *DLQ) Record(_ context.Context, message string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Path == "" {
		f.Path = "dlq.log"
	}
	entry := fmt.Sprintf("%s\n%s\n%s\n%s\n", fileSeparator, time.Now().Format(time.RFC3339), message, fileSeparator)
	return appendFile(f.Path, []byte(entry))
}

// Register registers the file provider.
func Register(r *dlq.Registry) {
	r.RegisterProvider("file", func(cfg config.RuntimeConfig, _ *dbpkg.Queries) dlq.DLQ {
		return &DLQ{Path: cfg.DLQFile}
	})
}
