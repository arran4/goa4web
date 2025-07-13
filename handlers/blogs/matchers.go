package blogs

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
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
		queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
		session, err := core.GetSession(r)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetBlogEntryForUserById(r.Context(), db.GetBlogEntryForUserByIdParams{
			ViewerIdusers: uid,
			ID:            int32(blogID),
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
		cd, _ := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
		if cd != nil && cd.HasRole("administrator") {
			ctx := context.WithValue(r.Context(), hcommon.KeyBlogEntry, row)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if cd == nil || !cd.HasRole("writer") || row.UsersIdusers != uid {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), hcommon.KeyBlogEntry, row)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
