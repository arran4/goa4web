package common

import (
	"net/http"

	searchutil "github.com/arran4/goa4web/workers/searchworker"
)

// SearchWords returns the cached search terms for the current request.
// A copy is returned to prevent accidental mutation of CoreData state.
func (cd *CoreData) SearchWords() []string {
	if cd == nil {
		return nil
	}
	return append([]string(nil), cd.cache.searchWords...)
}

func (cd *CoreData) searchWordsFromRequest(r *http.Request) []string {
	if cd.cache.searchWords != nil {
		return cd.cache.searchWords
	}
	words := searchutil.BreakupTextToWords(r.PostFormValue("searchwords"))
	if len(words) == 0 {
		cd.cache.searchWords = []string{}
		return cd.cache.searchWords
	}
	cd.cache.searchWords = append([]string(nil), words...)
	return cd.cache.searchWords
}
