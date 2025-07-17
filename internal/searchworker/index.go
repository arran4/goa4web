package searchworker

import (
	"context"
	"strings"

	db "github.com/arran4/goa4web/internal/db"
)

// SearchWordIDs parses text into unique words, inserts them and returns their IDs.
func SearchWordIDs(ctx context.Context, text string, q *db.Queries) ([]int64, error) {
	words := map[string]struct{}{}
	for _, w := range BreakupTextToWords(text) {
		words[strings.ToLower(w)] = struct{}{}
	}
	ids := make([]int64, 0, len(words))
	for w := range words {
		id, err := q.CreateSearchWord(ctx, strings.ToLower(w))
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// InsertWords executes insertFn for each word ID.
func InsertWordsCtx(ctx context.Context, wordIDs []int64, insertFn InsertFunc) error {
	for _, wid := range wordIDs {
		if err := insertFn(ctx, wid); err != nil {
			return err
		}
	}
	return nil
}

// InsertWordsToForumSearchCtx associates words with a forum comment.
func InsertWordsToForumSearchCtx(ctx context.Context, wordIDs []int64, q *db.Queries, cid int64) error {
	return InsertWordsCtx(ctx, wordIDs, func(c context.Context, wid int64) error {
		return q.AddToForumCommentSearch(c, db.AddToForumCommentSearchParams{
			CommentID:                      int32(cid),
			SearchwordlistIdsearchwordlist: int32(wid),
		})
	})
}

// InsertWordsToImageSearchCtx associates words with an image post.
func InsertWordsToImageSearchCtx(ctx context.Context, wordIDs []int64, q *db.Queries, pid int64) error {
	return InsertWordsCtx(ctx, wordIDs, func(c context.Context, wid int64) error {
		return q.AddToImagePostSearch(c, db.AddToImagePostSearchParams{
			ImagePostID:                    int32(pid),
			SearchwordlistIdsearchwordlist: int32(wid),
		})
	})
}

// InsertWordsToWritingSearchCtx associates words with a writing post.
func InsertWordsToWritingSearchCtx(ctx context.Context, wordIDs []int64, q *db.Queries, wid int64) error {
	return InsertWordsCtx(ctx, wordIDs, func(c context.Context, id int64) error {
		return q.AddToForumWritingSearch(c, db.AddToForumWritingSearchParams{
			WritingID:                      int32(wid),
			SearchwordlistIdsearchwordlist: int32(id),
		})
	})
}

// InsertWordsToLinkerSearchCtx associates words with a linker entry.
func InsertWordsToLinkerSearchCtx(ctx context.Context, wordIDs []int64, q *db.Queries, lid int64) error {
	return InsertWordsCtx(ctx, wordIDs, func(c context.Context, wid int64) error {
		return q.AddToLinkerSearch(c, db.AddToLinkerSearchParams{
			LinkerID:                       int32(lid),
			SearchwordlistIdsearchwordlist: int32(wid),
		})
	})
}
