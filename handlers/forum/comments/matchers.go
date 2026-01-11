package comments

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
)

// RequireCommentAuthor ensures the requester authored the comment referenced in the URL.
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
		session, err := core.GetSession(r)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		uid, _ := session.Values["UID"].(int32)

		authorized := row.UsersIdusers == uid
		if !authorized && cd != nil {
			if cd.IsAdmin() {
				authorized = true
			} else {
				authorized = cd.HasGrant("forum", "thread", "edit-any", row.ForumthreadID) ||
					cd.HasGrant("forum", "thread", "edit", row.ForumthreadID)
			}
		}
		if !authorized {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
