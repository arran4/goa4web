package search

import (
	"context"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/workers/searchworker"
)

func indexText(ctx context.Context, q *db.Queries, text string, add func(context.Context, int64, int32) error) error {
	counts := map[string]int32{}
	for _, w := range searchworker.BreakupTextToWords(text) {
		counts[strings.ToLower(w)]++
	}
	for w, c := range counts {
		id, err := q.CreateSearchWord(ctx, w)
		if err != nil {
			return err
		}
		if err := add(ctx, id, c); err != nil {
			return err
		}
	}
	return nil
}
