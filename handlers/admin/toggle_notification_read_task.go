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

// ToggleNotificationReadTask switches notification read state.
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
		n, err := queries.GetNotification(r.Context(), int32(id))
		if err != nil {
			return fmt.Errorf("load notification fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if n.ReadAt.Valid {
			err = queries.MarkNotificationUnread(r.Context(), int32(id))
		} else {
			err = queries.MarkNotificationRead(r.Context(), int32(id))
		}
		if err != nil {
			return fmt.Errorf("toggle read fail %w", handlers.ErrRedirectOnSamePageHandler(err))
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
		return "toggled notifications " + ids + " read state"
	}
	return "toggled notification read state"
}
