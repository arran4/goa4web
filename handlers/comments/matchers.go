package comments

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
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
		queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
		session, err := core.GetSession(r)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetCommentByIdForUser(r.Context(), db.GetCommentByIdForUserParams{
			UsersIdusers: uid,
			Idcomments:   int32(commentID),
		})
		if err != nil {
			log.Printf("Error: %s", err)
			http.NotFound(w, r)
			return
		}

		cd, _ := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
		if row.UsersIdusers != uid && (cd == nil || !cd.HasRole("administrator")) {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), hcommon.KeyComment, row)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
