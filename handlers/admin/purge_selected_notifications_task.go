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

// PurgeSelectedNotificationsTask removes selected notifications.
type PurgeSelectedNotificationsTask struct{ tasks.TaskString }

var purgeSelectedNotificationsTask = &PurgeSelectedNotificationsTask{TaskString: TaskPurgeSelected}

// ensures PurgeSelectedNotificationsTask implements the tasks.Task interface
var _ tasks.Task = (*PurgeSelectedNotificationsTask)(nil)
var _ tasks.AuditableTask = (*PurgeSelectedNotificationsTask)(nil)

func (PurgeSelectedNotificationsTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, idStr := range r.Form["id"] {
		if idStr == "" {
			continue
		}
		id, _ := strconv.Atoi(idStr)
		if err := queries.DeleteNotification(r.Context(), int32(id)); err != nil {
			return fmt.Errorf("delete notification %w", handlers.ErrRedirectOnSamePageHandler(err))
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

// AuditRecord summarises notifications being deleted.
func (PurgeSelectedNotificationsTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["DeletedID"].(string); ok && ids != "" {
		return "deleted notifications " + ids
	}
	return "deleted notifications"
}
