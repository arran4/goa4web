package searchworker

import (
	"context"
	"log"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

// IndexedTask describes a task that can be indexed by the search worker.
// It exposes metadata about the indexed item and the text snippets to add to
// the search tables.
type IndexedTask interface {
	// IndexType returns the search index type to update.
	IndexType() string
	// IndexData extracts the pieces of data to index from the event payload.
	// Returning nil indicates there is nothing to index.
	IndexData(data map[string]any) []IndexEventData
}

// processEvent indexes text for tasks implementing IndexableTask.
func processEvent(ctx context.Context, evt eventbus.TaskEvent, q *dbpkg.Queries) {
	task, ok := evt.Task.(IndexedTask)
	if !ok || evt.Data == nil {
		return
	}
	typ := task.IndexType()
	data := task.IndexData(evt.Data)
	if typ == "" || len(data) == 0 {
		return
	}

	for _, d := range data {
		if d.Type == "" {
			d.Type = typ
		}
		if d.ID == 0 || d.Text == "" {
			continue
		}
		if err := index(ctx, q, d); err != nil {
			log.Printf("index error: %v", err)
		}
	}
}
