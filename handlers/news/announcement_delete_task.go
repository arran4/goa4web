package news

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// AnnouncementDeleteTask disables the announcement for a news post.
type AnnouncementDeleteTask struct{ tasks.TaskString }

var announcementDeleteTask = &AnnouncementDeleteTask{TaskString: TaskDelete}

var _ tasks.Task = (*AnnouncementDeleteTask)(nil)

func (AnnouncementDeleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	pid, _ := strconv.Atoi(mux.Vars(r)["news"])
	if err := cd.DeleteAnnouncement(int32(pid)); err != nil {
		return fmt.Errorf("delete announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
