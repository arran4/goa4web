package writings

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// WritingAuthor ensures the requester authored the writing referenced in the URL.
func WritingAuthor() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		vars := mux.Vars(r)
		writingID, _ := strconv.Atoi(vars["writing"])
		queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
		session, err := core.GetSession(r)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), db.GetWritingByIdForUserDescendingByPublishedDateParams{
			Userid:    uid,
			Idwriting: int32(writingID),
		})
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.UsersIdusers == uid
	}
}
