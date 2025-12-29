package searchworker

import (
	"context"
	"log"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/stats"
)

// EventKey is the map key used for search index events.
const EventKey = "search_index"

// Index types handled by the worker.
const (
	TypeComment = "comment"
	TypeWriting = "writing"
	TypeLinker  = "linker"
	TypeImage   = "image"
)

// IndexEventData describes content to index.
type IndexEventData struct {
	Type string
	ID   int32
	Text string
}

// Worker listens for index events and updates the search tables.
func Worker(ctx context.Context, bus *eventbus.Bus, q db.Querier) {
	if bus == nil || q == nil {
		return
	}
	ch := bus.Subscribe(eventbus.TaskMessageType)
	for {
		select {
		case msg := <-ch:
			evt, ok := msg.(eventbus.TaskEvent)
			if !ok {
				continue
			}
			// Use a context that isn't canceled while processing
			// the current event so database operations aren't
			// interrupted mid-flight when the worker context is
			// cancelled.
			evtCtx := context.WithoutCancel(ctx)
			if data, ok := evt.Data[EventKey].(IndexEventData); ok {
				if err := index(evtCtx, q, data); err != nil {
					log.Printf("index error: %v", err)
				}
			} else {
				processEvent(evtCtx, evt, q)
			}
		case <-ctx.Done():
			return
		}
	}
}

func index(ctx context.Context, q db.Querier, data IndexEventData) error {
	counts := map[string]int32{}
	for _, w := range BreakupTextToWords(data.Text) {
		counts[strings.ToLower(w)]++
	}
	for word, count := range counts {
		id, err := q.SystemCreateSearchWord(ctx, strings.ToLower(word))
		if err != nil {
			return err
		}
		switch data.Type {
		case TypeComment:
			if err := q.SystemAddToForumCommentSearch(ctx, db.SystemAddToForumCommentSearchParams{CommentID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		case TypeWriting:
			if err := q.SystemAddToForumWritingSearch(ctx, db.SystemAddToForumWritingSearchParams{WritingID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		case TypeLinker:
			if err := q.SystemAddToLinkerSearch(ctx, db.SystemAddToLinkerSearchParams{LinkerID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		case TypeImage:
			if err := q.SystemAddToImagePostSearch(ctx, db.SystemAddToImagePostSearchParams{ImagePostID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		}
	}
	switch data.Type {
	case TypeComment:
		if err := q.SystemSetCommentLastIndex(ctx, data.ID); err != nil {
			return err
		}
	case TypeWriting:
		if err := q.SystemSetWritingLastIndex(ctx, data.ID); err != nil {
			return err
		}
	case TypeLinker:
		if err := q.SystemSetLinkerLastIndex(ctx, data.ID); err != nil {
			return err
		}
	case TypeImage:
		if err := q.SystemSetImagePostLastIndex(ctx, data.ID); err != nil {
			return err
		}
	}
	stats.Inc("indexed_items")
	return nil
}
