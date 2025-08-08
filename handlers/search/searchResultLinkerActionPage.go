package search

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type SearchLinkerTask struct{ tasks.TaskString }

var searchLinkerTask = &SearchLinkerTask{TaskString: TaskSearchLinker}
var _ tasks.Task = (*SearchLinkerTask)(nil)

func (SearchLinkerTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !common.CanSearch(cd, "linker") {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return nil
	}
	if err := cd.SearchLinker(r); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return nil
	}
	return handlers.TemplateWithDataHandler("resultLinkerActionPage.gohtml", struct{}{})
}
