package searchworker

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	searchdomain "github.com/arran4/goa4web/internal/search"
)

func isAlphanumericOrPunctuation(char rune) bool {
	return searchdomain.IsTokenRune(char)
}

// IsAlphanumericOrPunctuation is exported for testing.
func IsAlphanumericOrPunctuation(char rune) bool {
	return isAlphanumericOrPunctuation(char)
}

// BreakupTextToWords splits input into tokens of alphanumeric or
// punctuation characters used for search indexing.
func BreakupTextToWords(input string) []string {
	return searchdomain.BreakupTextToWords(input)
}

// SearchWordIdsFromText inserts new search words and returns their ids.
// It redirects on error and returns true when a redirect has been issued.
type WordCount struct {
	ID    int64
	Count int32
}

func SearchWordIdsFromText(w http.ResponseWriter, r *http.Request, text string, queries db.Querier) ([]WordCount, bool) {
	counts := map[string]int32{}
	for _, word := range BreakupTextToWords(text) {
		counts[strings.ToLower(word)]++
	}
	results := make([]WordCount, 0, len(counts))
	for word, c := range counts {
		id, err := queries.SystemCreateSearchWord(r.Context(), strings.ToLower(word))
		if err != nil {
			log.Printf("Error: createSearchWord: %s", err)
			redirectSeeOtherWithError(w, r, err)
			return nil, true
		}
		results = append(results, WordCount{ID: id, Count: c})
	}
	return results, false
}

// InsertWordsToLinkerSearch associates search words with a linker post.
func InsertWordsToLinkerSearch(w http.ResponseWriter, r *http.Request, words []WordCount, queries db.Querier, lid int64) bool {
	return InsertWords(w, r, words, func(ctx context.Context, wid int64, count int32) error {
		return queries.SystemAddToLinkerSearch(ctx, db.SystemAddToLinkerSearchParams{
			LinkerID:                       int32(lid),
			SearchwordlistIdsearchwordlist: int32(wid),
			WordCount:                      count,
		})
	})
}

// InsertWordsToImageSearch associates search words with an image post.
func InsertWordsToImageSearch(w http.ResponseWriter, r *http.Request, words []WordCount, queries db.Querier, pid int64) bool {
	return InsertWords(w, r, words, func(ctx context.Context, wid int64, count int32) error {
		return queries.SystemAddToImagePostSearch(ctx, db.SystemAddToImagePostSearchParams{
			ImagePostID:                    int32(pid),
			SearchwordlistIdsearchwordlist: int32(wid),
			WordCount:                      count,
		})
	})
}

// InsertWordsToWritingSearch associates search words with a writing post.
func InsertWordsToWritingSearch(w http.ResponseWriter, r *http.Request, words []WordCount, queries db.Querier, wacid int64) bool {
	return InsertWords(w, r, words, func(ctx context.Context, wid int64, count int32) error {
		return queries.SystemAddToForumWritingSearch(ctx, db.SystemAddToForumWritingSearchParams{
			WritingID:                      int32(wacid),
			SearchwordlistIdsearchwordlist: int32(wid),
			WordCount:                      count,
		})
	})
}

// InsertWordsToForumSearch associates search words with a forum comment.
func InsertWordsToForumSearch(w http.ResponseWriter, r *http.Request, words []WordCount, queries db.Querier, cid int64) bool {
	return InsertWords(w, r, words, func(ctx context.Context, wid int64, count int32) error {
		return queries.SystemAddToForumCommentSearch(ctx, db.SystemAddToForumCommentSearchParams{
			CommentID:                      int32(cid),
			SearchwordlistIdsearchwordlist: int32(wid),
			WordCount:                      count,
		})
	})
}
