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

// AnnouncementAddTask enables an announcement for a news post.
type AnnouncementAddTask struct{ tasks.TaskString }

var announcementAddTask = &AnnouncementAddTask{TaskString: TaskAdd}

var _ tasks.Task = (*AnnouncementAddTask)(nil)

func (AnnouncementAddTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	pid, _ := strconv.Atoi(mux.Vars(r)["news"])
	if err := cd.AddAnnouncement(int32(pid)); err != nil {
		return fmt.Errorf("add announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
