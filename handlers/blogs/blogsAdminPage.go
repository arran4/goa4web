package blogs

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// AdminPage shows the blog administration index with a list of blogs.
func AdminPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	offset := cd.Offset()
	ps := cd.PageSize()
	cd.NextLink = fmt.Sprintf("/admin/blogs?offset=%d", offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("/admin/blogs?offset=%d", offset-ps)
		cd.StartLink = "/admin/blogs?offset=0"
	}
	cd.PageTitle = "Blog Admin"
	BlogsAdminPageTmpl.Handle(w, r, struct{}{})
}

const BlogsAdminPageTmpl handlers.Page = "blogs/adminPage.gohtml"
