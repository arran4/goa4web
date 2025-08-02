package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// PurgeReadNotificationsTask removes all read notifications.
type PurgeReadNotificationsTask struct{ tasks.TaskString }

var purgeReadNotificationsTask = &PurgeReadNotificationsTask{TaskString: TaskPurgeRead}

var _ tasks.Task = (*PurgeReadNotificationsTask)(nil)
var _ tasks.AuditableTask = (*PurgeReadNotificationsTask)(nil)

func (PurgeReadNotificationsTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := queries.AdminPurgeReadNotifications(r.Context()); err != nil {
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

func (PurgeReadNotificationsTask) AuditRecord(map[string]any) string {
	return "purged read notifications"
}
