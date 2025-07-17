package news

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	db "github.com/arran4/goa4web/internal/db"
)

// RequireNewsPostAuthor ensures the requester authored the news post referenced in the URL.
func RequireNewsPostAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.Atoi(mux.Vars(r)["post"])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		queries := r.Context().Value(corecorecommon.KeyQueries).(*db.Queries)
		session, err := core.GetSession(r)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetForumThreadIdByNewsPostId(r.Context(), int32(postID))
		if err != nil {
			log.Printf("Error: %s", err)
			http.NotFound(w, r)
			return
		}

		if row.Idusers.Int32 != uid {
			http.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), corecorecommon.KeyNewsPost, row)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
