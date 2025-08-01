package news

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// AnnouncementAddTask enables an announcement for a news post.
type AnnouncementAddTask struct{ tasks.TaskString }

var announcementAddTask = &AnnouncementAddTask{TaskString: TaskAdd}

var _ tasks.Task = (*AnnouncementAddTask)(nil)

func (AnnouncementAddTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])

	ann, err := cd.NewsAnnouncement(int32(pid))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("get announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	if ann == nil {
		if err := queries.AdminPromoteAnnouncement(r.Context(), int32(pid)); err != nil {
			return fmt.Errorf("create announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	} else if !ann.Active {
		if err := queries.SetAnnouncementActive(r.Context(), db.SetAnnouncementActiveParams{Active: true, ID: ann.ID}); err != nil {
			return fmt.Errorf("activate announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return nil
}
