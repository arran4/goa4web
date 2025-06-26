package search

import (
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
