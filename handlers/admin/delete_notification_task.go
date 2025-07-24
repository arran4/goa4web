package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteNotificationTask deletes notifications entirely.
type DeleteNotificationTask struct{ tasks.TaskString }

var deleteNotificationTask = &DeleteNotificationTask{TaskString: TaskDeleteNotification}

var _ tasks.Task = (*DeleteNotificationTask)(nil)
var _ tasks.AuditableTask = (*DeleteNotificationTask)(nil)

func (DeleteNotificationTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.DeleteNotification(r.Context(), int32(id)); err != nil {
			return fmt.Errorf("delete notification fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["DeletedID"] = appendID(evt.Data["DeletedID"], id)
			}
		}
	}
	return nil
}

func (DeleteNotificationTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["DeletedID"].(string); ok {
		return "deleted notifications " + ids
	}
	return "deleted notifications"
}
