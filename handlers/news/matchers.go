package news

import (
	"context"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// RequireNewsPostAuthor ensures the requester authored the news post referenced in the URL.
func RequireNewsPostAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.Atoi(mux.Vars(r)["post"])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		queries := cd.Queries()
		session, err := cd.GetSession(r)
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

		ctx := context.WithValue(r.Context(), consts.KeyNewsPost, row)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
