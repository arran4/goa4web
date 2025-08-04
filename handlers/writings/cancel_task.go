package writings

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// CancelTask cancels comment editing.
type CancelTask struct{ tasks.TaskString }

var cancelTask = &CancelTask{TaskString: TaskCancel}

// CancelTask is only used to abort editing, implementing tasks.Task ensures it fits the routing interface even though no additional behaviour is required.
var _ tasks.Task = (*CancelTask)(nil)

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	return handlers.RedirectHandler("/writings/article/" + mux.Vars(r)["writing"])
}
