package writings

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// RequireWritingAuthor ensures the requester authored the writing referenced in the URL.
func RequireWritingAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writing, err := cd.CurrentWriting()
		if err != nil {
			log.Printf("RequireWritingAuthor load writing: %v", err)
			http.NotFound(w, r)
			return
		}
		if writing == nil {
			http.NotFound(w, r)
			return
		}
		if cd.HasAdminRole() {
			next.ServeHTTP(w, r)
			return
		}
		if !cd.HasContentWriterRole() || writing.UsersIdusers != cd.UserID {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
