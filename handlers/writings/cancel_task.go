package writings

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// CancelTask cancels comment editing.
type CancelTask struct{ tasks.TaskString }

var cancelTask = &CancelTask{TaskString: TaskCancel}

// CancelTask is only used to abort editing, implementing tasks.Task ensures it fits the routing interface even though no additional behaviour is required.
var _ tasks.Task = (*CancelTask)(nil)

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.CurrentWriting()
	if err != nil || writing == nil {
		return fmt.Errorf("load writing fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("/writings/article/%d", writing.Idwriting))
}
