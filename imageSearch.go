package goa4web

import (
	"log"
	"net/http"
)

func InsertWordsToImageSearch(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *Queries, pid int64) bool {
	for _, wid := range wordIds {
		if err := queries.AddToImagePostSearch(r.Context(), AddToImagePostSearchParams{
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
