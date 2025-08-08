package search

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchBlogsTask struct{ tasks.TaskString }

var searchBlogsTask = &SearchBlogsTask{TaskString: TaskSearchBlogs}
var _ tasks.Task = (*SearchBlogsTask)(nil)

func (SearchBlogsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !common.CanSearch(cd, "blogs") {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return nil
	}
	if err := cd.SearchBlogs(r); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return nil
	}
	return handlers.TemplateWithDataHandler("resultBlogsActionPage.gohtml", struct{}{})
}
