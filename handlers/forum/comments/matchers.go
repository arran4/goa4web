package comments

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
)

// RequireCommentAuthor ensures the requester authored the comment referenced in the URL.
func RequireCommentAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commentID, err := strconv.Atoi(mux.Vars(r)["comment"])
		if err != nil {
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

		row, err := queries.GetCommentByIdForUser(r.Context(), db.GetCommentByIdForUserParams{
			ViewerID: uid,
			ID:       int32(commentID),
			UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			log.Printf("Error: %s", err)
			http.NotFound(w, r)
			return
		}

		cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if row.UsersIdusers != uid && (cd == nil || !cd.HasRole("administrator")) {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), consts.KeyComment, row)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
