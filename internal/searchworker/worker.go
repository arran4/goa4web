package searchworker

import (
	"context"
	"log"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

const (
	SearchForumComment = "forum_comment"
)

// BusWorker listens for search events on the bus and updates search tables accordingly.
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
	task, ok := evt.Task.(IndexableTask)
	if !ok {
		return
	}
	if evt.Data == nil {
		return
	}
	typ := task.IndexType()
	text := task.IndexText(evt.Data)
	id := task.IndexID(evt.Data)
	if typ == "" || text == "" || id == 0 {
		return
	}

	wordIDs, err := searchWordIDs(ctx, q, text)
	if err != nil {
		log.Printf("search worker tokenize: %v", err)
		return
	}
	switch typ {
	case SearchForumComment:
		insertWords(ctx, wordIDs, func(ctx context.Context, wid int64) error {
			return q.AddToForumCommentSearch(ctx, db.AddToForumCommentSearchParams{
				CommentID:                      int32(id),
				SearchwordlistIdsearchwordlist: int32(wid),
			})
		})
	}
}

func searchWordIDs(ctx context.Context, q *db.Queries, text string) ([]int64, error) {
	words := map[string]struct{}{}
	for _, w := range BreakupTextToWords(text) {
		words[strings.ToLower(w)] = struct{}{}
	}
	ids := make([]int64, 0, len(words))
	for w := range words {
		id, err := q.CreateSearchWord(ctx, w)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func insertWords(ctx context.Context, wordIDs []int64, fn func(context.Context, int64) error) {
	for _, wid := range wordIDs {
		if err := fn(ctx, wid); err != nil {
			log.Printf("insert search word: %v", err)
		}
	}
}
