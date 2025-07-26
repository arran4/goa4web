package writings

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/lazy"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
)

// RequireWritingAuthor ensures the requester authored the writing referenced in the URL.
func RequireWritingAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		writingIDStr := vars["article"]
		if writingIDStr == "" {
			writingIDStr = vars["writing"]
		}
		writingID, err := strconv.Atoi(writingIDStr)
		if err != nil {
			log.Printf("RequireWritingAuthor invalid writing ID %q: %v", writingIDStr, err)
			http.NotFound(w, r)
			return
		}
		queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
		session, err := core.GetSession(r)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetWritingByIdForUserDescendingByPublishedDate(r.Context(), db.GetWritingByIdForUserDescendingByPublishedDateParams{
			ViewerID:      uid,
			Idwriting:     int32(writingID),
			ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			log.Printf("Error: %s", err)
			http.NotFound(w, r)
			return
		}

		cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd != nil {
			cd.WritingByID(int32(writingID), lazy.Set[*db.GetWritingByIdForUserDescendingByPublishedDateRow](row))
			cd.SetCurrentWriting(int32(writingID))
		}
		if cd != nil && cd.HasAdminRole() {
			next.ServeHTTP(w, r)
			return
		}
		if cd == nil || !cd.HasContentWriterRole() || row.UsersIdusers != uid {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
