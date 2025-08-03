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

var _ tasks.Task = (*ToggleNotificationReadTask)(nil)
var _ tasks.AuditableTask = (*ToggleNotificationReadTask)(nil)

func (ToggleNotificationReadTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		n, err := queries.AdminGetNotification(r.Context(), int32(id))
		if err != nil {
			return fmt.Errorf("get notification fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if n.ReadAt.Valid {
			if err := queries.AdminMarkNotificationUnread(r.Context(), int32(id)); err != nil {
				return fmt.Errorf("mark unread fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		} else {
			if err := queries.AdminMarkNotificationRead(r.Context(), int32(id)); err != nil {
				return fmt.Errorf("mark read fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["ToggledID"] = appendID(evt.Data["ToggledID"], id)
			}
		}
	}
	return nil
}

func (ToggleNotificationReadTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["ToggledID"].(string); ok {
		return "toggled notifications " + ids
	}
	return "toggled notifications"
}
