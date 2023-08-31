package main

import (
	"log"
	"net/http"
)

func InsertWordsToWritingSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *Queries, wacid int64) bool {
	for _, wid := range wordIds {
		if err := queries.addToForumWritingSearch(r.Context(), addToForumWritingSearchParams{
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
