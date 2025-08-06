package blogs

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/lazy"
	"github.com/gorilla/mux"
)

// AdminBlogEditPage renders the edit form for a blog entry on the admin dashboard.
func AdminBlogEditPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	blogID, err := strconv.Atoi(vars["blog"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	queries := cd.Queries()
	row, err := queries.GetBlogEntryForListerByID(r.Context(), db.GetBlogEntryForListerByIDParams{
		ListerID: cd.UserID,
		ID:       int32(blogID),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Blog not found"))
		return
	}
	cd.BlogEntryByID(int32(blogID), lazy.Set[*db.GetBlogEntryForListerByIDRow](row))
	cd.SetCurrentBlog(int32(blogID))
	cd.PageTitle = "Admin Edit Blog"
	if _, err := cd.Languages(); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	type Data struct {
		Blog    *db.GetBlogEntryForListerByIDRow
		Mode    string
		PostURL string
	}
	data := Data{
		Blog:    row,
		Mode:    "Edit",
		PostURL: fmt.Sprintf("/blogs/blog/%d/edit", blogID),
	}
	handlers.TemplateHandler(w, r, "blogsAdminBlogEditPage.gohtml", data)
}
