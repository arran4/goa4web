package file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
)

// DLQ appends messages to a file.
type DLQ struct {
	Path     string
	mu       sync.Mutex
	Appender appender
}

// fileSeparator marks the boundary around each recorded message.
const fileSeparator = "-----"

type appender interface {
	Append(name string, data []byte) error
}

type osAppender struct{}

func (osAppender) Append(name string, data []byte) error {
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
	app := f.Appender
	if app == nil {
		app = osAppender{}
	}
	return app.Append(f.Path, []byte(entry))
}

// Register registers the file provider.
func Register(r *dlq.Registry) {
	r.RegisterProvider("file", func(cfg *config.RuntimeConfig, _ *dbpkg.Queries) dlq.DLQ {
		return &DLQ{Path: cfg.DLQFile, Appender: osAppender{}}
	})
}

// Record represents a DLQ entry stored in a file.
type Record struct {
	Time    time.Time
	Message string
}

// List reads path and returns up to limit records in newest-first order.
func List(path string, limit int) ([]Record, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	var recs []Record
	s := bufio.NewScanner(fh)
	for s.Scan() {
		if s.Text() != fileSeparator {
			continue
		}
		if !s.Scan() {
			break
		}
		t, err := time.Parse(time.RFC3339, s.Text())
		if err != nil {
			break
		}
		if !s.Scan() {
			break
		}
		msg := s.Text()
		if !s.Scan() {
			break
		}
		recs = append(recs, Record{Time: t, Message: msg})
		if limit > 0 && len(recs) >= limit {
			break
		}
	}
	sort.Slice(recs, func(i, j int) bool { return recs[i].Time.After(recs[j].Time) })
	if err := s.Err(); err != nil {
		return recs, err
	}
	return recs, nil
}
