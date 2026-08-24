package comments

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core"
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
		session, err := core.GetSession(r)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		uid, _ := session.Values["UID"].(int32)

		authorized := false
		if cd != nil {
			if cd.IsAdmin() {
				authorized = true
			} else {
				section := cd.Section()
				if section == "" {
					section = "forum" // fallback
				}

				// If they own the comment, they need 'edit'. Otherwise they need 'edit-any'.
				if row.UsersIdusers == uid {
					authorized = cd.HasGrant(section, "thread", "edit", row.ForumthreadID) ||
					    cd.HasGrant(section, "comment", "edit", row.Idcomments) ||
						cd.HasGrant(section, "topic", "edit", cd.SelectedThreadID())
				} else {
					authorized = cd.HasGrant(section, "thread", "edit-any", row.ForumthreadID) ||
					    cd.HasGrant(section, "comment", "edit-any", row.Idcomments) ||
						cd.HasGrant(section, "topic", "edit-any", cd.SelectedThreadID())
				}
			}
		}

		if !authorized {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
