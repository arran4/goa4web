package writings

import (
	"log"
	"net/http"
	"strings"
	"unicode"

	db "github.com/arran4/goa4web/internal/db"
)

func isAlphanumericOrPunctuation(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) || strings.ContainsRune("'-", char)
}

func breakupTextToWords(input string) []string {
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
	for _, word := range breakupTextToWords(text) {
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
