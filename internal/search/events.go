// Package search owns search-index event contracts and operations.
package search

import (
	"context"
	"strings"
	"unicode"

	"github.com/arran4/goa4web/internal/db"
)

// IndexEventKey identifies search-index requests in an event payload.
const IndexEventKey = "search_index"

// Search index types handled by Index.
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

// IsTokenRune reports whether char belongs in a search token.
func IsTokenRune(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) || strings.ContainsRune("'-", char)
}

// BreakupTextToWords splits input into tokens used for search indexing.
func BreakupTextToWords(input string) []string {
	var tokens []string
	startIndex := -1
	for i, char := range input {
		if IsTokenRune(char) {
			if startIndex == -1 {
				startIndex = i
			}
		} else if startIndex != -1 {
			tokens = append(tokens, input[startIndex:i])
			startIndex = -1
		}
	}
	if startIndex != -1 {
		tokens = append(tokens, input[startIndex:])
	}
	return tokens
}

// Index synchronously updates the search index for one item.
func Index(ctx context.Context, q db.Querier, data IndexEventData) error {
	counts := map[string]int32{}
	for _, word := range BreakupTextToWords(data.Text) {
		counts[strings.ToLower(word)]++
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
		return q.SystemSetCommentLastIndex(ctx, data.ID)
	case TypeWriting:
		return q.SystemSetWritingLastIndex(ctx, data.ID)
	case TypeLinker:
		return q.SystemSetLinkerLastIndex(ctx, data.ID)
	case TypeImage:
		return q.SystemSetImagePostLastIndex(ctx, data.ID)
	}
	return nil
}
