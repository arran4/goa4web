package blogs

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
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
	row, err := cd.EditableBlogPost(int32(blogID))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Blog not found"))
		return
	}
	cd.BlogEntryByID(int32(blogID), lazy.Set(row))
	cd.SetCurrentBlog(int32(blogID))
	cd.PageTitle = "Admin Edit Blog"
	if _, err := cd.Languages(); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	labels, _ := cd.BlogAuthorLabels(int32(blogID))
	type Data struct {
		Mode         string
		PostURL      string
		AuthorLabels []string
	}
	data := Data{
		Mode:         "Edit",
		PostURL:      fmt.Sprintf("/blogs/blog/%d/edit", blogID),
		AuthorLabels: labels,
	}
	BlogsAdminBlogEditPageTmpl.Handle(w, r, data)
}

const BlogsAdminBlogEditPageTmpl handlers.Page = "blogs/blogsAdminBlogEditPage.gohtml"
