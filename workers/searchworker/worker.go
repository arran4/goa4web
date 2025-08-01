package searchworker

import (
	"context"
	"log"
	"strings"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
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
func Worker(ctx context.Context, bus *eventbus.Bus, q *dbpkg.Queries) {
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

func index(ctx context.Context, q *dbpkg.Queries, data IndexEventData) error {
	counts := map[string]int32{}
	for _, w := range BreakupTextToWords(data.Text) {
		counts[strings.ToLower(w)]++
	}
	for word, count := range counts {
		id, err := q.CreateSearchWord(ctx, strings.ToLower(word))
		if err != nil {
			return err
		}
		switch data.Type {
		case TypeComment:
			if err := q.AddToForumCommentSearch(ctx, dbpkg.AddToForumCommentSearchParams{CommentID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		case TypeWriting:
			if err := q.AddToForumWritingSearch(ctx, dbpkg.AddToForumWritingSearchParams{WritingID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		case TypeLinker:
			if err := q.AddToLinkerSearch(ctx, dbpkg.AddToLinkerSearchParams{LinkerID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		case TypeImage:
			if err := q.AddToImagePostSearch(ctx, dbpkg.AddToImagePostSearchParams{ImagePostID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		}
	}
	switch data.Type {
	case TypeComment:
		return q.SetCommentLastIndexForSystem(ctx, data.ID)
	case TypeWriting:
		return q.SetWritingLastIndex(ctx, data.ID)
	case TypeLinker:
		return q.SetLinkerLastIndex(ctx, data.ID)
	case TypeImage:
		return q.SetImagePostLastIndex(ctx, data.ID)
	}
	return nil
}
