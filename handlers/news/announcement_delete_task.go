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

// AnnouncementDeleteTask disables the announcement for a news post.
type AnnouncementDeleteTask struct{ tasks.TaskString }

var announcementDeleteTask = &AnnouncementDeleteTask{TaskString: TaskDelete}

var _ tasks.Task = (*AnnouncementDeleteTask)(nil)

func (AnnouncementDeleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasRole("administrator") {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}

	queries := cd.Queries()
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])

	ann, err := cd.NewsAnnouncement(int32(pid))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("announcement for news fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		return nil
	}
	if ann != nil && ann.Active {
		if err := queries.SetAnnouncementActive(r.Context(), db.SetAnnouncementActiveParams{Active: false, ID: ann.ID}); err != nil {
			return fmt.Errorf("deactivate announcement fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return nil
}
