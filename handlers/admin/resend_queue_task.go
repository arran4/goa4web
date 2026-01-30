package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/emailqueue"
)

// ResendQueueTask triggers sending queued emails immediately.
type ResendQueueTask struct{ tasks.TaskString }

var resendQueueTask = &ResendQueueTask{TaskString: TaskResend}

// ensure ResendQueueTask satisfies the tasks.Task interface
var _ tasks.Task = (*ResendQueueTask)(nil)
var _ tasks.AuditableTask = (*ResendQueueTask)(nil)

func (ResendQueueTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	provider := cd.EmailProvider()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	selection := r.Form.Get("selection")
	scope := "ids"
	var ids []int32
	var emails []*db.AdminGetPendingEmailByIDRow
	if selection == "filtered" {
		scope = "filtered"
		langID, role := emailFiltersFromRequest(r)
		filterPrefix := ""
		if strings.Contains(r.URL.Path, "/failed") {
			filterPrefix = "failed"
			rows, err := queries.AdminListFailedEmailIDs(r.Context(), db.AdminListFailedEmailIDsParams{
				LanguageID: langID,
				RoleName:   role,
			})
			if err != nil {
				return fmt.Errorf("list failed email ids fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			for _, id := range rows {
				ids = append(ids, id)
			}
		} else {
			rows, err := queries.AdminListUnsentPendingEmails(r.Context(), db.AdminListUnsentPendingEmailsParams{
				LanguageID: langID,
				RoleName:   role,
			})
			if err != nil {
				return fmt.Errorf("list pending emails fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			for _, row := range rows {
				ids = append(ids, row.ID)
				emails = append(emails, &db.AdminGetPendingEmailByIDRow{
					ID:          row.ID,
					ToUserID:    row.ToUserID,
					Body:        row.Body,
					ErrorCount:  row.ErrorCount,
					DirectEmail: row.DirectEmail,
				})
			}
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["QueuedEmailCount"] = len(ids)
				evt.Data["QueuedEmailFilter"] = emailFilterSummary(filterPrefix, langID, role)
			}
		}
	} else {
		for _, idStr := range r.Form["id"] {
			id, _ := strconv.Atoi(idStr)
			ids = append(ids, int32(id))
			e, err := queries.AdminGetPendingEmailByID(r.Context(), int32(id))
			if err != nil {
				return fmt.Errorf("get email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			emails = append(emails, e)
		}
	}
	if len(emails) == 0 && len(ids) > 0 {
		for _, id := range ids {
			e, err := queries.AdminGetPendingEmailByID(r.Context(), id)
			if err != nil {
				return fmt.Errorf("get email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			emails = append(emails, e)
		}
	}
	for _, e := range emails {
		addr, err := emailqueue.ResolveQueuedEmailAddress(r.Context(), queries, cd.Config, &db.SystemListPendingEmailsRow{ID: e.ID, ToUserID: e.ToUserID, Body: e.Body, ErrorCount: e.ErrorCount, DirectEmail: e.DirectEmail})
		if err != nil {
			return fmt.Errorf("resolve address fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if provider != nil {
			if err := provider.Send(r.Context(), addr, []byte(e.Body)); err != nil {
				return fmt.Errorf("send email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
		if err := queries.SystemMarkPendingEmailSent(r.Context(), e.ID); err != nil {
			return fmt.Errorf("mark sent fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if selection != "filtered" {
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
				if evt := cd.Event(); evt != nil {
					if evt.Data == nil {
						evt.Data = map[string]any{}
					}
					evt.Data["QueuedEmailID"] = appendID(evt.Data["QueuedEmailID"], int(e.ID))
				}
			}
		}
	}
	return buildEmailTaskRedirect(r, "resent", scope, ids)
}

// AuditRecord summarises queued emails being resent.
func (ResendQueueTask) AuditRecord(data map[string]any) string {
	if count, ok := data["QueuedEmailCount"]; ok {
		summary := "resent queued emails (" + fmt.Sprint(count) + ")"
		if filter, ok := data["QueuedEmailFilter"].(string); ok && filter != "" {
			summary += " " + filter
		}
		return summary
	}
	if ids, ok := data["QueuedEmailID"].(string); ok {
		return "resent queued emails " + ids
	}
	return "resent queued emails"
}
