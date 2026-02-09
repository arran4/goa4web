package blogs

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/go-be-lazy"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// RequireBlogAuthor ensures the requester authored the blog entry referenced in the URL.
func RequireBlogAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		cd.LoadSelectionsFromRequest(r)

		vars := mux.Vars(r)
		blogID, err := strconv.Atoi(vars["blog"])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		queries := cd.Queries()
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
				handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			}
			return
		}
		if cd != nil {
			cd.BlogEntryByID(int32(blogID), lazy.Set[*db.GetBlogEntryForListerByIDRow](row))
			cd.SetCurrentBlog(int32(blogID))
		}
		if cd == nil {
			http.NotFound(w, r)
			return
		}
		hasEditGrant := cd.HasGrant("blogs", "entry", "edit", row.Idblogs)
		hasEditAnyGrant := cd.HasGrant("blogs", "entry", "edit-any", 0)
		if !(hasEditGrant || hasEditAnyGrant) {
			http.NotFound(w, r)
			return
		}
		if !hasEditAnyGrant && row.UsersIdusers != uid {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireBlogEditGrant checks whether the requester can edit the blog referenced in the URL path.
func RequireBlogEditGrant() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd == nil {
			return false
		}
		vars := mux.Vars(r)
		blogID, err := strconv.Atoi(vars["blog"])
		if err != nil {
			return false
		}
		if cd.HasGrant("blogs", "entry", "edit-any", 0) {
			return true
		}
		return cd.HasGrant("blogs", "entry", "edit", int32(blogID))
	}
}

// RequireBlogCommentAccess ensures the requester has view or reply access to the blog entry.
func RequireBlogCommentAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		cd.LoadSelectionsFromRequest(r)

		blog, err := cd.BlogPost()
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			} else {
				log.Printf("BlogPost: %v", err)
				handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			}
			return
		}

		if _, err := cd.BlogCommentThread(); err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("BlogCommentThread: %v", err)
		}

		if !(cd.HasGrant("blogs", "entry", "view", blog.Idblogs) ||
			cd.HasGrant("blogs", "entry", "reply", blog.Idblogs) ||
			cd.SelectedThreadCanReply()) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
