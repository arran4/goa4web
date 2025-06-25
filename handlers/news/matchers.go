package news

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// NewsPostAuthor ensures the requester authored the news post referenced in the URL.
func NewsPostAuthor() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		vars := mux.Vars(r)
		postID, _ := strconv.Atoi(vars["post"])
		queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
		session, err := core.GetSession(r)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetForumThreadIdByNewsPostId(r.Context(), int32(postID))
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.Idusers.Int32 == uid
	}
}
