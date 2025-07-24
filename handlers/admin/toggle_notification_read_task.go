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

// ToggleNotificationReadTask toggles the read state of a notification.
type ToggleNotificationReadTask struct{ tasks.TaskString }

var toggleNotificationReadTask = &ToggleNotificationReadTask{TaskString: TaskToggleRead}

// ensures ToggleNotificationReadTask implements the tasks.Task interface
var _ tasks.Task = (*ToggleNotificationReadTask)(nil)
var _ tasks.AuditableTask = (*ToggleNotificationReadTask)(nil)

func (ToggleNotificationReadTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	notif, err := queries.GetNotification(r.Context(), int32(id))
	if err != nil {
		return fmt.Errorf("get notification fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if notif.ReadAt.Valid {
		if err := queries.MarkNotificationUnread(r.Context(), int32(id)); err != nil {
			return fmt.Errorf("mark unread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	} else {
		if err := queries.MarkNotificationRead(r.Context(), int32(id)); err != nil {
			return fmt.Errorf("mark read fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["ToggledID"] = id
		}
	}
	return nil
}

// AuditRecord summarises the notification read toggle.
func (ToggleNotificationReadTask) AuditRecord(data map[string]any) string {
	if id, ok := data["ToggledID"].(int); ok {
		return fmt.Sprintf("toggled notification %d read state", id)
	}
	return "toggled notification read state"
}
