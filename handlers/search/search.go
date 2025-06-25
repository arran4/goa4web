package search

import (
	"database/sql"
	"errors"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strings"
	"unicode"
)

func isAlphanumericOrPunctuation(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) || strings.ContainsRune("'-", char)
}

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

func InsertWordsToLinkerSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, lid int64) bool {
	for _, wid := range wordIds {
		if err := queries.AddToLinkerSearch(r.Context(), db.AddToLinkerSearchParams{
			LinkerIdlinker:                 int32(lid),
			SearchwordlistIdsearchwordlist: int32(wid),
		}); err != nil {
			log.Printf("Error: addToLinkerSearch: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return true
		}
	}
	return false
}

func InsertWordsToImageSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, pid int64) bool {
	for _, wid := range wordIds {
		if err := queries.AddToImagePostSearch(r.Context(), db.AddToImagePostSearchParams{
			ImagepostIdimagepost:           int32(pid),
			SearchwordlistIdsearchwordlist: int32(wid),
		}); err != nil {
			log.Printf("Error: addToImagePostSearch: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return true
		}
	}
	return false
}

func InsertWordsToWritingSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, wacid int64) bool {
	for _, wid := range wordIds {
		if err := queries.AddToForumWritingSearch(r.Context(), db.AddToForumWritingSearchParams{
			WritingIdwriting:               int32(wacid),
			SearchwordlistIdsearchwordlist: int32(wid),
		}); err != nil {
			log.Printf("Error: addToForumWritingSearch: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return true
		}
	}
	return false
}

func InsertWordsToForumSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, cid int64) bool {
	for _, wid := range wordIds {
		if err := queries.AddToForumCommentSearch(r.Context(), db.AddToForumCommentSearchParams{
			CommentsIdcomments:             int32(cid),
			SearchwordlistIdsearchwordlist: int32(wid),
		}); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("Error: addToForumCommentSearch: %s", err)
				http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
				return true
			}
		}
	}
	return false
}
