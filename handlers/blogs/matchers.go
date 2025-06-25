package blogs

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// BlogAuthor ensures the requester authored the blog entry referenced in the URL.
func BlogAuthor() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		vars := mux.Vars(r)
		blogID, _ := strconv.Atoi(vars["blog"])
		queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
		session, err := core.GetSession(r)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetBlogEntryForUserById(r.Context(), int32(blogID))
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("Error: %s", err)
				return false
			}
		}

		return row.UsersIdusers == uid
	}
}
