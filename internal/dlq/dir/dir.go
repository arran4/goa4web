package dir

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
	id := ksuid.New()
	name := id.String() + ".txt"
	path := filepath.Join(d.Dir, name)

	now := time.Now()
	// RFC 5322 format
	header := fmt.Sprintf("Date: %s\nMessage-ID: <%s@dlq.local>\n\n",
		now.Format(time.RFC1123Z),
		id.String(),
	)

	return os.WriteFile(path, []byte(header+message+"\n"), 0o644)
}

// Get reads the message content for the given ID.
func (d *DLQ) Get(_ context.Context, id string) (string, error) {
	if d.Dir == "" {
		return "", os.ErrNotExist
	}
	safeName := filepath.Base(id)
	path := filepath.Join(d.Dir, safeName)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	// Check for RFC 5322 headers
	parts := strings.SplitN(content, "\n\n", 2)
	if len(parts) == 2 && (strings.Contains(parts[0], "Date: ") || strings.Contains(parts[0], "Message-ID: ")) {
		return strings.TrimSpace(parts[1]), nil
	}
	return strings.TrimSpace(content), nil
}

// Delete removes the message file for the given ID.
func (d *DLQ) Delete(_ context.Context, id string) error {
	if d.Dir == "" {
		return nil
	}
	safeName := filepath.Base(id)
	path := filepath.Join(d.Dir, safeName)
	return os.Remove(path)
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
	Size    int64
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

		size := int64(len(data))
		if info, err := e.Info(); err == nil {
			size = info.Size()
		}

		content := string(data)
		// Check for RFC 5322 headers
		parts := strings.SplitN(content, "\n\n", 2)
		var msg string
		if len(parts) == 2 && (strings.Contains(parts[0], "Date: ") || strings.Contains(parts[0], "Message-ID: ")) {
			msg = parts[1]
		} else {
			msg = content
		}

		recs = append(recs, Record{Name: e.Name(), Message: strings.TrimSpace(msg), Size: size})
		if limit > 0 && len(recs) >= limit {
			break
		}
	}
	return recs, nil
}
