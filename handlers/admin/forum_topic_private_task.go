package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// ConvertTopicToPrivateTask switches a forum topic to the private handler.
type ConvertTopicToPrivateTask struct{ tasks.TaskString }

var convertTopicToPrivateTask = &ConvertTopicToPrivateTask{TaskString: TaskForumTopicConvertPrivate}

var _ tasks.Task = (*ConvertTopicToPrivateTask)(nil)

func (ConvertTopicToPrivateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	tidStr := r.PostFormValue("topic_id")
	tid, err := strconv.Atoi(tidStr)
	if err != nil {
		return fmt.Errorf("topic id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.Queries().SystemSetForumTopicHandlerByID(r.Context(), db.SystemSetForumTopicHandlerByIDParams{
		Handler: "private",
		ID:      int32(tid),
	}); err != nil {
		return fmt.Errorf("update topic handler %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/maintenance"}
}
