package searchworker

import (
	"context"
	"log"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

// IndexRequest describes a search indexing operation to perform.
type IndexRequest struct {
	Type string
	ID   int64
	Text string
}

const (
	IndexForum   = "forum"
	IndexImage   = "image"
	IndexWriting = "writing"
	IndexLinker  = "linker"
)

// BusWorker listens for IndexRequest events and updates search tables.
func BusWorker(ctx context.Context, bus *eventbus.Bus, q *db.Queries) {
	if bus == nil || q == nil {
		return
	}
	ch := bus.Subscribe()
	for {
		select {
		case evt := <-ch:
			processEvent(ctx, q, evt)
		case <-ctx.Done():
			return
		}
	}
}

func processEvent(ctx context.Context, q *db.Queries, evt eventbus.Event) {
	if evt.Data == nil {
		return
	}
	v, ok := evt.Data["index"]
	if !ok {
		return
	}
	req, ok := v.(IndexRequest)
	if !ok {
		return
	}
	if err := processIndex(ctx, q, req); err != nil {
		log.Printf("search index: %v", err)
	}
}

func processIndex(ctx context.Context, q *db.Queries, req IndexRequest) error {
	ids, err := SearchWordIDs(ctx, req.Text, q)
	if err != nil {
		return err
	}
	switch req.Type {
	case IndexForum:
		return InsertWordsToForumSearchCtx(ctx, ids, q, req.ID)
	case IndexImage:
		return InsertWordsToImageSearchCtx(ctx, ids, q, req.ID)
	case IndexWriting:
		return InsertWordsToWritingSearchCtx(ctx, ids, q, req.ID)
	case IndexLinker:
		return InsertWordsToLinkerSearchCtx(ctx, ids, q, req.ID)
	default:
		return nil
	}
}
