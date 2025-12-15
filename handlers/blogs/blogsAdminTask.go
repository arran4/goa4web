package blogs

import (
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
)

type blogsAdminTask struct {
}

const (
	BlogsAdminPageTmpl = "blogs/adminPage.gohtml"
)

func NewBlogsAdminTask() tasks.Task {
	return &blogsAdminTask{}
}

func (t *blogsAdminTask) TemplatesRequired() []string {
	return []string{BlogsAdminPageTmpl}
}

func (t *blogsAdminTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *blogsAdminTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	offset := cd.Offset()
	ps := cd.PageSize()
	cd.NextLink = fmt.Sprintf("/admin/blogs?offset=%d", offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("/admin/blogs?offset=%d", offset-ps)
		cd.StartLink = "/admin/blogs?offset=0"
	}
	cd.PageTitle = "Blog Admin"
	if err := cd.ExecuteSiteTemplate(w, r, BlogsAdminPageTmpl, struct{}{}); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
