package admin

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/workers/emailqueue"
)

// ResendQueueTask triggers sending queued emails immediately.
type ResendQueueTask struct{ tasks.TaskString }

var resendQueueTask = &ResendQueueTask{TaskString: TaskResend}

// ensure ResendQueueTask satisfies the tasks.Task interface
var _ tasks.Task = (*ResendQueueTask)(nil)
var _ tasks.AuditableTask = (*ResendQueueTask)(nil)

func (ResendQueueTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	provider := email.ProviderFromConfig(config.AppRuntimeConfig)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	var emails []*db.GetPendingEmailByIDRow
	var ids []int32
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		e, err := queries.GetPendingEmailByID(r.Context(), int32(id))
		if err != nil {
			return fmt.Errorf("get email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		emails = append(emails, e)
		if e.ToUserID.Valid {
			ids = append(ids, e.ToUserID.Int32)
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["QueuedEmailID"] = appendID(evt.Data["QueuedEmailID"], id)
			}
		}
	}
	users := make(map[int32]*db.GetUserByIdRow)
	for _, id := range ids {
		if u, err := queries.GetUserById(r.Context(), id); err == nil {
			users[id] = u
		}
	}
	for _, e := range emails {
		addr, err := emailqueue.ResolveQueuedEmailAddress(r.Context(), queries, &db.FetchPendingEmailsRow{ID: e.ID, ToUserID: e.ToUserID, Body: e.Body, ErrorCount: e.ErrorCount, DirectEmail: e.DirectEmail}, config.AppRuntimeConfig)
		if err != nil {
			return fmt.Errorf("resolve address fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if provider != nil {
			if err := provider.Send(r.Context(), addr, []byte(e.Body)); err != nil {
				return fmt.Errorf("send email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
		if err := queries.MarkEmailSent(r.Context(), e.ID); err != nil {
			return fmt.Errorf("mark sent fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return nil
}

// AuditRecord summarises queued emails being resent.
func (ResendQueueTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["QueuedEmailID"].(string); ok {
		return "resent queued emails " + ids
	}
	return "resent queued emails"
}
