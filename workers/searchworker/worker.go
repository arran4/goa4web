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
	ch := bus.Subscribe()
	for {
		select {
		case evt := <-ch:
			if data, ok := evt.Data[EventKey].(IndexEventData); ok {
				if err := index(ctx, q, data); err != nil {
					log.Printf("index error: %v", err)
				}
			} else {
				processEvent(ctx, evt, q)
			}
		case <-ctx.Done():
			return
		}
	}
}

func index(ctx context.Context, q *dbpkg.Queries, data IndexEventData) error {
	words := map[string]struct{}{}
	for _, w := range BreakupTextToWords(data.Text) {
		words[strings.ToLower(w)] = struct{}{}
	}
	wordIDs := make([]int64, 0, len(words))
	for w := range words {
		id, err := q.CreateSearchWord(ctx, strings.ToLower(w))
		if err != nil {
			return err
		}
		wordIDs = append(wordIDs, id)
	}
	for _, wid := range wordIDs {
		switch data.Type {
		case TypeComment:
			if err := q.AddToForumCommentSearch(ctx, dbpkg.AddToForumCommentSearchParams{CommentID: data.ID, SearchwordlistIdsearchwordlist: int32(wid)}); err != nil {
				return err
			}
		case TypeWriting:
			if err := q.AddToForumWritingSearch(ctx, dbpkg.AddToForumWritingSearchParams{WritingID: data.ID, SearchwordlistIdsearchwordlist: int32(wid)}); err != nil {
				return err
			}
		case TypeLinker:
			if err := q.AddToLinkerSearch(ctx, dbpkg.AddToLinkerSearchParams{LinkerID: data.ID, SearchwordlistIdsearchwordlist: int32(wid)}); err != nil {
				return err
			}
		case TypeImage:
			if err := q.AddToImagePostSearch(ctx, dbpkg.AddToImagePostSearchParams{ImagePostID: data.ID, SearchwordlistIdsearchwordlist: int32(wid)}); err != nil {
				return err
			}
		}
	}
	switch data.Type {
	case TypeComment:
		return q.SetCommentLastIndex(ctx, data.ID)
	case TypeWriting:
		return q.SetWritingLastIndex(ctx, data.ID)
	case TypeLinker:
		return q.SetLinkerLastIndex(ctx, data.ID)
	case TypeImage:
		return q.SetImagePostLastIndex(ctx, data.ID)
	}
	return nil
}
