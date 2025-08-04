package writings

import (
	"fmt"
	"net/http"
	"strconv"

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
	vars := mux.Vars(r)
	writingID, err := strconv.Atoi(vars["writing"])
	if err != nil {
		return fmt.Errorf("writing id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RedirectHandler(fmt.Sprintf("/writings/article/%d", writingID))
}
