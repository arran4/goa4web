package search

import (
	"context"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/workers/searchworker"
)

// indexText tokenizes text, ensures the words exist in searchwordlist and then
// calls add for each (word id, count) pair. It caches word ids to avoid
// repeated database lookups when the same word occurs multiple times across
// documents.
func indexText(ctx context.Context, q *db.Queries, cache map[string]int64, text string, add func(context.Context, int64, int32) error) error {
	counts := map[string]int32{}
	for _, w := range searchworker.BreakupTextToWords(text) {
		counts[strings.ToLower(w)]++
	}
	for w, c := range counts {
		id, ok := cache[w]
		if !ok {
			var err error
			id, err = q.SystemCreateSearchWord(ctx, w)
			if err != nil {
				return err
			}
			if id == 0 {
				sw, err := q.SystemGetSearchWordByWordLowercased(ctx, w)
				if err != nil {
					return err
				}
				id = int64(sw.Idsearchwordlist)
			}
			cache[w] = id
		}
		if err := add(ctx, id, c); err != nil {
			return err
		}
	}
	return nil
}
