package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// threadNewCancelTask aborts creating a new thread.
type threadNewCancelTask struct{ tasks.TaskString }

var threadNewCancelAction = &threadNewCancelTask{TaskString: forumcommon.TaskCancel}

// ThreadNewCancelHandler aborts creating a new thread. Exported for reuse.
var ThreadNewCancelHandler = threadNewCancelAction

var _ tasks.Task = (*threadNewCancelTask)(nil)

func (threadNewCancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	topicID, _ := strconv.Atoi(mux.Vars(r)["topic"])
	base := "/forum"
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if cd.ForumBasePath != "" {
			base = cd.ForumBasePath
		}
	}
	return handlers.RedirectHandler(fmt.Sprintf("%s/topic/%d", base, topicID))
}
