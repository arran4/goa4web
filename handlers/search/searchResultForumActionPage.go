package search

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchForumTask struct{ tasks.TaskString }

var searchForumTask = &SearchForumTask{TaskString: TaskSearchForum}
var _ tasks.Task = (*SearchForumTask)(nil)

func (SearchForumTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !common.CanSearch(cd, "forum") {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return nil
	}
	if err := cd.SearchForum(r); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return nil
	}
	return handlers.TemplateWithDataHandler("resultForumActionPage.gohtml", struct{}{})
}
