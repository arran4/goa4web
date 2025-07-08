package searchutil

import (
	"context"
	"log"
	"net/http"
	"strings"
	"unicode"

	db "github.com/arran4/goa4web/internal/db"
)

func isAlphanumericOrPunctuation(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) || strings.ContainsRune("'-", char)
}

// IsAlphanumericOrPunctuation is exported for testing.
func IsAlphanumericOrPunctuation(char rune) bool {
	return isAlphanumericOrPunctuation(char)
}

// BreakupTextToWords splits input into tokens of alphanumeric or
// punctuation characters used for search indexing.
func BreakupTextToWords(input string) []string {
	var tokens []string
	startIndex := -1
	for i, char := range input {
		if isAlphanumericOrPunctuation(char) {
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

// SearchWordIdsFromText inserts new search words and returns their ids.
// It redirects on error and returns true when a redirect has been issued.
func SearchWordIdsFromText(w http.ResponseWriter, r *http.Request, text string, queries *db.Queries) ([]int64, bool) {
	words := map[string]int32{}
	for _, word := range BreakupTextToWords(text) {
		words[strings.ToLower(word)] = 0
	}
	wordIds := make([]int64, 0, len(words))
	for word := range words {
		id, err := queries.CreateSearchWord(r.Context(), strings.ToLower(word))
		if err != nil {
			log.Printf("Error: createSearchWord: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return nil, true
		}
		wordIds = append(wordIds, id)
	}
	return wordIds, false
}

// InsertWordsToLinkerSearch associates search words with a linker post.
func InsertWordsToLinkerSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, lid int64) bool {
	return InsertWords(w, r, wordIds, func(ctx context.Context, wid int64) error {
		return queries.AddToLinkerSearch(ctx, db.AddToLinkerSearchParams{
			LinkerIdlinker:                 int32(lid),
			SearchwordlistIdsearchwordlist: int32(wid),
		})
	})
}

// InsertWordsToImageSearch associates search words with an image post.
func InsertWordsToImageSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, pid int64) bool {
	return InsertWords(w, r, wordIds, func(ctx context.Context, wid int64) error {
		return queries.AddToImagePostSearch(ctx, db.AddToImagePostSearchParams{
			ImagepostIdimagepost:           int32(pid),
			SearchwordlistIdsearchwordlist: int32(wid),
		})
	})
}

// InsertWordsToWritingSearch associates search words with a writing post.
func InsertWordsToWritingSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, wacid int64) bool {
	return InsertWords(w, r, wordIds, func(ctx context.Context, wid int64) error {
		return queries.AddToForumWritingSearch(ctx, db.AddToForumWritingSearchParams{
			WritingID:                      int32(wacid),
			SearchwordlistIdsearchwordlist: int32(wid),
		})
	})
}

// InsertWordsToForumSearch associates search words with a forum comment.
func InsertWordsToForumSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, cid int64) bool {
	return InsertWords(w, r, wordIds, func(ctx context.Context, wid int64) error {
		return queries.AddToForumCommentSearch(ctx, db.AddToForumCommentSearchParams{
			CommentsIdcomments:             int32(cid),
			SearchwordlistIdsearchwordlist: int32(wid),
		})
	})
}
