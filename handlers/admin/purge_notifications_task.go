package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// PurgeNotificationsTask removes old read notifications.
type PurgeNotificationsTask struct{ tasks.TaskString }

var purgeNotificationsTask = &PurgeNotificationsTask{TaskString: TaskPurge}

// ensures PurgeNotificationsTask implements the tasks.Task interface
var _ tasks.Task = (*PurgeNotificationsTask)(nil)
var _ tasks.AuditableTask = (*PurgeNotificationsTask)(nil)

func (PurgeNotificationsTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := queries.PurgeReadNotifications(r.Context()); err != nil {
		return fmt.Errorf("purge notifications fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Purged"] = true
		}
	}
	return nil
}

// AuditRecord summarises purging notifications.
func (PurgeNotificationsTask) AuditRecord(map[string]any) string {
	return "purged read notifications"
}
