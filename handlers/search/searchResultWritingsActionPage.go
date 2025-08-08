package search

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchWritingsTask struct{ tasks.TaskString }

var searchWritingsTask = &SearchWritingsTask{TaskString: TaskSearchWritings}
var _ tasks.Task = (*SearchWritingsTask)(nil)

func (SearchWritingsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !common.CanSearch(cd, "writing") {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return nil
	}
	if err := cd.SearchWritings(r); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return nil
	}
	return handlers.TemplateWithDataHandler("resultWritingsActionPage.gohtml", struct{}{})
}
