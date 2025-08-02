package blogs

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/lazy"
)

// RequireBlogAuthor ensures the requester authored the blog entry referenced in the URL.
func RequireBlogAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		blogID, err := strconv.Atoi(vars["blog"])
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

		row, err := queries.GetBlogEntryForListerByID(r.Context(), db.GetBlogEntryForListerByIDParams{
			ListerID: uid,
			ID:       int32(blogID),
			UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				http.NotFound(w, r)
			default:
				log.Printf("Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd != nil {
			cd.BlogEntryByID(int32(blogID), lazy.Set[*db.GetBlogEntryForListerByIDRow](row))
			cd.SetCurrentBlog(int32(blogID))
		}
		if cd != nil && cd.HasRole("administrator") {
			next.ServeHTTP(w, r)
			return
		}
		if cd == nil || !cd.HasRole("writer") || row.UsersIdusers != uid {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
