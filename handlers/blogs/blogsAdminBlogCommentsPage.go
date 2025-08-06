package blogs

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminBlogCommentsPage lists comments for a blog entry.
func AdminBlogCommentsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Blog     *db.GetBlogEntryForListerByIDRow
		Comments []*db.GetCommentsByThreadIdForUserRow
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	blogID, err := strconv.Atoi(vars["blog"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	queries := cd.Queries()
	blog, err := queries.GetBlogEntryForListerByID(r.Context(), db.GetBlogEntryForListerByIDParams{
		ListerID: cd.UserID,
		ID:       int32(blogID),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Blog not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Blog %d Comments Admin", blog.Idblogs)
	data := Data{Blog: blog}
	if blog.ForumthreadID.Valid {
		if rows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
			ViewerID: cd.UserID,
			ThreadID: blog.ForumthreadID.Int32,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		}); err == nil {
			data.Comments = rows
		}
	}
	handlers.TemplateHandler(w, r, "blogsAdminBlogCommentsPage.gohtml", data)
}
