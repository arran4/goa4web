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
	"github.com/arran4/goa4web/workers/emailqueue"
)

// ResendSentEmailTask re-sends already delivered emails immediately.
type ResendSentEmailTask struct{ tasks.TaskString }

var resendSentEmailTask = &ResendSentEmailTask{TaskString: TaskResend}

// ensure ResendSentEmailTask satisfies the tasks.Task interface
var _ tasks.Task = (*ResendSentEmailTask)(nil)
var _ tasks.AuditableTask = (*ResendSentEmailTask)(nil)

func (ResendSentEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	provider := cd.EmailProvider()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		e, err := queries.AdminGetPendingEmailByID(r.Context(), int32(id))
		if err != nil {
			return fmt.Errorf("get email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		addr, err := emailqueue.ResolveQueuedEmailAddress(r.Context(), queries, cd.Config, &db.FetchPendingEmailsRow{ID: e.ID, ToUserID: e.ToUserID, Body: e.Body, ErrorCount: e.ErrorCount, DirectEmail: e.DirectEmail})
		if err != nil {
			return fmt.Errorf("resolve address fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if provider != nil {
			if err := provider.Send(r.Context(), addr, []byte(e.Body)); err != nil {
				return fmt.Errorf("send email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["SentEmailID"] = appendID(evt.Data["SentEmailID"], id)
			}
		}
	}
	return nil
}

// AuditRecord summarises sent emails being resent.
func (ResendSentEmailTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["SentEmailID"].(string); ok {
		return "resent sent emails " + ids
	}
	return "resent sent emails"
}
