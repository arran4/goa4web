package dir

import (
	"context"
	"os"
	"path/filepath"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/segmentio/ksuid"
)

// DLQ writes each message to a new file inside Dir using a KSUID filename.
type DLQ struct {
	Dir string
}

// Record writes the message to a unique file within the directory.
func (d *DLQ) Record(_ context.Context, message string) error {
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

// Register registers the directory provider.
func Register(r *dlq.Registry) {
	r.RegisterProvider("dir", func(cfg config.RuntimeConfig, _ *dbpkg.Queries) dlq.DLQ {
		return &DLQ{Dir: cfg.DLQFile}
	})
}
