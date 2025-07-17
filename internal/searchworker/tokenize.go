package searchworker

import (
	"context"
	"log"
	"net/http"
	"strings"
	"unicode"

	db "github.com/arran4/goa4web/internal/db"
)

// TODO move all of this to searchworker

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
	ids, err := SearchWordIDs(r.Context(), text, queries)
	if err != nil {
		log.Printf("Error: createSearchWord: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return nil, true
	}
	return ids, false
}

// InsertWordsToLinkerSearch associates search words with a linker post.
func InsertWordsToLinkerSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, lid int64) bool {
	if err := InsertWordsToLinkerSearchCtx(r.Context(), wordIds, queries, lid); err != nil {
		log.Printf("insert linker search: %v", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return true
	}
	return false
}

// InsertWordsToImageSearch associates search words with an image post.
func InsertWordsToImageSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, pid int64) bool {
	if err := InsertWordsToImageSearchCtx(r.Context(), wordIds, queries, pid); err != nil {
		log.Printf("insert image search: %v", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return true
	}
	return false
}

// InsertWordsToWritingSearch associates search words with a writing post.
func InsertWordsToWritingSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, wacid int64) bool {
	if err := InsertWordsToWritingSearchCtx(r.Context(), wordIds, queries, wacid); err != nil {
		log.Printf("insert writing search: %v", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return true
	}
	return false
}

// InsertWordsToForumSearch associates search words with a forum comment.
func InsertWordsToForumSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, cid int64) bool {
	if err := InsertWordsToForumSearchCtx(r.Context(), wordIds, queries, cid); err != nil {
		log.Printf("insert forum search: %v", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return true
	}
	return false
}
