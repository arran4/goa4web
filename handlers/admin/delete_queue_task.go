package admin

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/handlers"
)

// DeleteQueueTask removes queued emails without sending.
type DeleteQueueTask struct{ tasks.TaskString }

var deleteQueueTask = &DeleteQueueTask{TaskString: TaskDelete}

// ensure DeleteQueueTask satisfies the tasks.Task interface
var _ tasks.Task = (*DeleteQueueTask)(nil)
var _ tasks.AuditableTask = (*DeleteQueueTask)(nil)

func (DeleteQueueTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.AdminDeletePendingEmail(r.Context(), int32(id)); err != nil {
			return fmt.Errorf("delete email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["DeletedEmailID"] = appendID(evt.Data["DeletedEmailID"], id)
			}
		}
	}
	return nil
}

// AuditRecord summarises queued emails being removed.
func (DeleteQueueTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["DeletedEmailID"].(string); ok {
		return "deleted queued emails " + ids
	}
	return "deleted queued emails"
}
