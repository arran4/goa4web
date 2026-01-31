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
	selection := r.Form.Get("selection")
	scope := "ids"
	var ids []int32
	if selection == "filtered" {
		scope = "filtered"
		filters := emailFiltersFromRequest(r)
		rows, err := queries.AdminListSentEmailIDs(r.Context(), db.AdminListSentEmailIDsParams{
			LanguageID:    filters.LangIDParam(),
			RoleName:      filters.Role,
			Provider:      filters.ProviderParam(),
			CreatedBefore: filters.CreatedBefore,
		})
		if err != nil {
			return fmt.Errorf("list sent email ids fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		for _, id := range rows {
			ids = append(ids, id)
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["SentEmailCount"] = len(ids)
				evt.Data["SentEmailFilter"] = emailFilterSummary("", filters)
			}
		}
	} else {
		for _, idStr := range r.Form["id"] {
			id, _ := strconv.Atoi(idStr)
			ids = append(ids, int32(id))
		}
	}
	for _, id := range ids {
		e, err := queries.AdminGetPendingEmailByID(r.Context(), id)
		if err != nil {
			return fmt.Errorf("get email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		addr, err := emailqueue.ResolveQueuedEmailAddress(r.Context(), queries, cd.Config, &db.SystemListPendingEmailsRow{ID: e.ID, ToUserID: e.ToUserID, Body: e.Body, ErrorCount: e.ErrorCount, DirectEmail: e.DirectEmail})
		if err != nil {
			return fmt.Errorf("resolve address fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if provider != nil {
			if err := provider.Send(r.Context(), addr, []byte(e.Body)); err != nil {
				return fmt.Errorf("send email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
		if selection != "filtered" {
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
				if evt := cd.Event(); evt != nil {
					if evt.Data == nil {
						evt.Data = map[string]any{}
					}
					evt.Data["SentEmailID"] = appendID(evt.Data["SentEmailID"], int(id))
				}
			}
		}
	}
	return buildEmailTaskRedirect(r, "resent", scope, ids)
}

// AuditRecord summarises sent emails being resent.
func (ResendSentEmailTask) AuditRecord(data map[string]any) string {
	if count, ok := data["SentEmailCount"]; ok {
		summary := "resent sent emails (" + fmt.Sprint(count) + ")"
		if filter, ok := data["SentEmailFilter"].(string); ok && filter != "" {
			summary += " " + filter
		}
		return summary
	}
	if ids, ok := data["SentEmailID"].(string); ok {
		return "resent sent emails " + ids
	}
	return "resent sent emails"
}
