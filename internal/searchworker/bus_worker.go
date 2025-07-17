package searchworker

import (
	"context"
	"log"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

// BusWorker listens for search events and updates the search indices.
func BusWorker(ctx context.Context, bus *eventbus.Bus, q *db.Queries) {
	if bus == nil || q == nil {
		return
	}
	ch := bus.Subscribe()
	for {
		select {
		case evt := <-ch:
			processEvent(ctx, evt, q)
		case <-ctx.Done():
			return
		}
	}
}

func processEvent(ctx context.Context, evt eventbus.Event, q *db.Queries) {
	if evt.Data == nil {
		return
	}
	text, ok := evt.Data["search_text"].(string)
	if !ok || text == "" {
		return
	}
	table, ok := evt.Data["search_table"].(string)
	if !ok {
		return
	}
	var id int64
	switch v := evt.Data["search_id"].(type) {
	case int64:
		id = v
	case int32:
		id = int64(v)
	case float64:
		id = int64(v)
	}

	words := map[string]struct{}{}
	for _, w := range BreakupTextToWords(text) {
		words[strings.ToLower(w)] = struct{}{}
	}

	for word := range words {
		wid, err := q.CreateSearchWord(ctx, strings.ToLower(word))
		if err != nil {
			log.Printf("create search word: %v", err)
			continue
		}
		switch table {
		case "forum":
			if err := q.AddToForumCommentSearch(ctx, db.AddToForumCommentSearchParams{
				CommentID:                      int32(id),
				SearchwordlistIdsearchwordlist: int32(wid),
			}); err != nil {
				log.Printf("add to forum search: %v", err)
			}
		case "image":
			if err := q.AddToImagePostSearch(ctx, db.AddToImagePostSearchParams{
				ImagePostID:                    int32(id),
				SearchwordlistIdsearchwordlist: int32(wid),
			}); err != nil {
				log.Printf("add to image search: %v", err)
			}
		case "writing":
			if err := q.AddToForumWritingSearch(ctx, db.AddToForumWritingSearchParams{
				WritingID:                      int32(id),
				SearchwordlistIdsearchwordlist: int32(wid),
			}); err != nil {
				log.Printf("add to writing search: %v", err)
			}
		case "linker":
			if err := q.AddToLinkerSearch(ctx, db.AddToLinkerSearchParams{
				LinkerID:                       int32(id),
				SearchwordlistIdsearchwordlist: int32(wid),
			}); err != nil {
				log.Printf("add to linker search: %v", err)
			}
		}
	}
}
