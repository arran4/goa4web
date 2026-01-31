package file

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/segmentio/ksuid"
)

// DLQ appends messages to a file.
type DLQ struct {
	Path     string
	mu       sync.Mutex
	Appender appender
}

// fileSeparator marks the boundary around each recorded message in legacy format.
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

	now := time.Now()
	id := ksuid.New().String()

	// Escape "From " at beginning of lines to be mbox compatible
	var body bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(message))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "From ") {
			body.WriteString(">" + line + "\n")
		} else {
			body.WriteString(line + "\n")
		}
	}

	// mbox format:
	// From <sender> <date>
	// Headers...
	// <blank line>
	// Body...
	// <blank line>

	header := fmt.Sprintf("From DLQ %s\nDate: %s\nMessage-ID: <%s@dlq.local>\nContent-Length: %d\n\n",
		now.Format(time.ANSIC),
		now.Format(time.RFC1123Z),
		id,
		body.Len(),
	)

	entry := header + body.String() + "\n"

	app := f.Appender
	if app == nil {
		app = osAppender{}
	}
	return app.Append(f.Path, []byte(entry))
}

// Register registers the file provider.
func Register(r *dlq.Registry) {
	r.RegisterProvider("file", func(cfg *config.RuntimeConfig, _ db.Querier) dlq.DLQ {
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

	scanner := bufio.NewScanner(fh)

	var (
		currentMsg  strings.Builder
		currentTime time.Time
		inLegacy    bool
		inMbox      bool
	)

	flushMbox := func() {
		if inMbox {
			recs = append(recs, Record{Time: currentTime, Message: strings.TrimSpace(currentMsg.String())})
			currentMsg.Reset()
			inMbox = false
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		if line == fileSeparator {
			if inLegacy {
				// End of legacy message
				recs = append(recs, Record{Time: currentTime, Message: strings.TrimSpace(currentMsg.String())})
				currentMsg.Reset()
				inLegacy = false
			} else {
				// Start of legacy message
				flushMbox() // Should not happen if strictly alternating but possible
				if scanner.Scan() {
					tStr := scanner.Text()
					t, err := time.Parse(time.RFC3339, tStr)
					if err == nil {
						currentTime = t
						inLegacy = true
					}
				}
			}
			continue
		}

		if strings.HasPrefix(line, "From DLQ ") {
			flushMbox()
			// Parse date from From line
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 3 {
				// ANSIC is "Mon Jan _2 15:04:05 2006"
				t, err := time.Parse(time.ANSIC, parts[2])
				if err == nil {
					currentTime = t
				}
			}
			inMbox = true

			// Read headers until blank line
			for scanner.Scan() {
				headerLine := scanner.Text()
				if headerLine == "" {
					break
				}
				if strings.HasPrefix(headerLine, "Date: ") {
					t, err := time.Parse(time.RFC1123Z, strings.TrimPrefix(headerLine, "Date: "))
					if err == nil {
						currentTime = t
					}
				}
			}
			continue
		}

		if inLegacy {
			currentMsg.WriteString(line + "\n")
		} else if inMbox {
			if strings.HasPrefix(line, ">From ") {
				line = line[1:]
			}
			currentMsg.WriteString(line + "\n")
		}
	}

	flushMbox()

	sort.Slice(recs, func(i, j int) bool { return recs[i].Time.After(recs[j].Time) })
	if err := scanner.Err(); err != nil {
		return recs, err
	}

	if limit > 0 && len(recs) > limit {
		return recs[:limit], nil
	}

	return recs, nil
}
