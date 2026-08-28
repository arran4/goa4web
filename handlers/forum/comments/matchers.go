package comments

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

// RequireCommentAuthor ensures the requester is authorized to edit the comment referenced in the URL.
func RequireCommentAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		row, err := cd.CurrentComment(r)
		if err != nil {
			log.Printf("Error: %s", err)

			http.NotFound(w, r)
			return
		}
		if row == nil {

			http.NotFound(w, r)
			return
		}

		authorized := false
		if cd != nil {

			authorized = cd.CanEditCommentTarget(row.Idcomments, row.ForumthreadID, row.UsersIdusers)
		}

		if !authorized {

			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
