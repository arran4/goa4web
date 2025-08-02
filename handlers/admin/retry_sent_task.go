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

// RetrySentEmailTask queues a previously sent email for delivery again.
type RetrySentEmailTask struct{ tasks.TaskString }

var retrySentEmailTask = &RetrySentEmailTask{TaskString: TaskRetry}

var _ tasks.Task = (*RetrySentEmailTask)(nil)
var _ tasks.AuditableTask = (*RetrySentEmailTask)(nil)

func (RetrySentEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		e, err := queries.AdminGetPendingEmailByID(r.Context(), int32(id))
		if err != nil {
			return fmt.Errorf("get email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if err := queries.InsertPendingEmail(r.Context(), db.InsertPendingEmailParams{ToUserID: e.ToUserID, Body: e.Body, DirectEmail: e.DirectEmail}); err != nil {
			return fmt.Errorf("queue email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["RetryEmailID"] = appendID(evt.Data["RetryEmailID"], id)
			}
		}
	}
	return nil
}

func (RetrySentEmailTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["RetryEmailID"].(string); ok {
		return "retried sent emails " + ids
	}
	return "retried sent emails"
}
