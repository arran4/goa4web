package forum

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// threadNewCancelTask aborts creating a new thread.
type threadNewCancelTask struct{ tasks.TaskString }

var threadNewCancelAction = &threadNewCancelTask{TaskString: TaskCancel}

var _ tasks.Task = (*threadNewCancelTask)(nil)

func (threadNewCancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	topicID, _ := strconv.Atoi(mux.Vars(r)["topic"])
	return handlers.RedirectHandler(fmt.Sprintf("/forum/topic/%d", topicID))
}
