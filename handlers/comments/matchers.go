package comments

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// Author ensures the requester authored the comment referenced in the URL.
func Author() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		vars := mux.Vars(r)
		commentID, _ := strconv.Atoi(vars["comment"])
		queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
		session, err := core.GetSession(r)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetCommentByIdForUser(r.Context(), db.GetCommentByIdForUserParams{
			UsersIdusers: uid,
			Idcomments:   int32(commentID),
		})
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.UsersIdusers == uid
	}
}
