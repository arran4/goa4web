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

// MarkReadTask marks notifications as read.
type MarkReadTask struct{ tasks.TaskString }

var markReadTask = &MarkReadTask{TaskString: TaskDismiss}

// ensures MarkReadTask implements the tasks.Task interface
var _ tasks.Task = (*MarkReadTask)(nil)
var _ tasks.AuditableTask = (*MarkReadTask)(nil)

func (MarkReadTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.MarkNotificationRead(r.Context(), int32(id)); err != nil {
			return fmt.Errorf("mark read fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["MarkedID"] = appendID(evt.Data["MarkedID"], id)
			}
		}
	}
	return nil
}

// AuditRecord summarises notifications being marked read.
func (MarkReadTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["MarkedID"].(string); ok {
		return "marked notifications " + ids + " read"
	}
	return "marked notifications read"
}
