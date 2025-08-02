package dir

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
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
	r.RegisterProvider("dir", func(cfg *config.RuntimeConfig, _ db.Querier) dlq.DLQ {
		return &DLQ{Dir: cfg.DLQFile}
	})
}

// Record represents a DLQ message stored in a directory file.
type Record struct {
	Name    string
	Message string
}

// List reads up to limit records from dir sorted by filename descending.
func List(dir string, limit int) ([]Record, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() > entries[j].Name() })
	var recs []Record
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		recs = append(recs, Record{Name: e.Name(), Message: strings.TrimSpace(string(data))})
		if limit > 0 && len(recs) >= limit {
			break
		}
	}
	return recs, nil
}
