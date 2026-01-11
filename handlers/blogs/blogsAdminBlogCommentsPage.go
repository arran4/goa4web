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
)

// AdminBlogCommentsPage lists comments for a blog entry.
func AdminBlogCommentsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	blogID, err := strconv.Atoi(vars["blog"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	cd.SetCurrentBlog(int32(blogID))
	blog, err := cd.BlogPost()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Blog not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Blog %d Comments Admin", blog.Idblogs)
	if _, err := cd.BlogCommentThread(); err != nil && err != sql.ErrNoRows {
		// ignore but log? There is no log imported; but we can ignore.
	}
	BlogsAdminBlogCommentsPageTmpl.Handle(w, r, struct{}{})
}

const BlogsAdminBlogCommentsPageTmpl handlers.Page = "blogs/blogsAdminBlogCommentsPage.gohtml"
