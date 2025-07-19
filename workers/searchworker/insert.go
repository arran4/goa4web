package searchworker

import (
	"context"
	"log"
	"net/http"
)

// InsertFunc performs an insert operation for a single search word.
type InsertFunc func(ctx context.Context, wordID int64) error

// InsertWords executes insertFn for each word ID. It redirects on error and
// returns true when a redirect has been issued.
func InsertWords(w http.ResponseWriter, r *http.Request, wordIDs []int64, insertFn InsertFunc) bool {
	for _, wid := range wordIDs {
		if err := insertFn(r.Context(), wid); err != nil {
			log.Printf("Error inserting search word: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return true
		}
	}
	return false
}
