package news

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/lazy"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
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

		cd.NewsPostByID(int32(postID), lazy.Set[*db.GetForumThreadIdByNewsPostIdRow](row))
		cd.SetCurrentNewsPost(int32(postID))
		next.ServeHTTP(w, r)
	})
}
