package admin

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/emailqueue"
)

// BulkResendQueueTask retries all filtered queued emails.
type BulkResendQueueTask struct{ tasks.TaskString }

var bulkResendQueueTask = &BulkResendQueueTask{TaskString: TaskResendFilteredQueue}

// ensure BulkResendQueueTask satisfies the tasks.Task interface
var _ tasks.Task = (*BulkResendQueueTask)(nil)
var _ tasks.AuditableTask = (*BulkResendQueueTask)(nil)

func (BulkResendQueueTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	provider := cd.EmailProvider()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	filters := emailFiltersFromValues(r.PostForm)
	rows, err := queries.AdminListUnsentPendingEmails(r.Context(), db.AdminListUnsentPendingEmailsParams{
		LanguageID:    filters.LangIDParam(),
		RoleName:      filters.Role,
		Status:        filters.StatusParam(),
		Provider:      filters.ProviderParam(),
		CreatedBefore: filters.CreatedBefore,
		Limit:         2147483647,
		Offset:        0,
	})
	if err != nil {
		return fmt.Errorf("list filtered emails fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["BulkQueuedEmailCount"] = len(rows)
			evt.Data["BulkQueuedEmailFilter"] = filters.AuditSummary()
		}
	}
	for _, e := range rows {
		addr, err := emailqueue.ResolveQueuedEmailAddress(r.Context(), queries, cd.Config, &db.SystemListPendingEmailsRow{
			ID:          e.ID,
			ToUserID:    e.ToUserID,
			Body:        e.Body,
			ErrorCount:  e.ErrorCount,
			DirectEmail: e.DirectEmail,
		})
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
	}
	return nil
}

// AuditRecord summarises retrying all filtered queued emails.
func (BulkResendQueueTask) AuditRecord(data map[string]any) string {
	count, _ := data["BulkQueuedEmailCount"].(int)
	if filter, ok := data["BulkQueuedEmailFilter"].(string); ok && filter != "" {
		return fmt.Sprintf("resent %d queued emails (%s)", count, filter)
	}
	if count > 0 {
		return fmt.Sprintf("resent %d queued emails", count)
	}
	return "resent queued emails"
}

// BulkDeleteQueueTask deletes all filtered queued emails.
type BulkDeleteQueueTask struct{ tasks.TaskString }

var bulkDeleteQueueTask = &BulkDeleteQueueTask{TaskString: TaskDeleteFilteredQueue}

// ensure BulkDeleteQueueTask satisfies the tasks.Task interface
var _ tasks.Task = (*BulkDeleteQueueTask)(nil)
var _ tasks.AuditableTask = (*BulkDeleteQueueTask)(nil)

func (BulkDeleteQueueTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	filters := emailFiltersFromValues(r.PostForm)
	rows, err := queries.AdminListUnsentPendingEmails(r.Context(), db.AdminListUnsentPendingEmailsParams{
		LanguageID:    filters.LangIDParam(),
		RoleName:      filters.Role,
		Status:        filters.StatusParam(),
		Provider:      filters.ProviderParam(),
		CreatedBefore: filters.CreatedBefore,
		Limit:         2147483647,
		Offset:        0,
	})
	if err != nil {
		return fmt.Errorf("list filtered emails fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, e := range rows {
		if err := queries.AdminDeletePendingEmail(r.Context(), e.ID); err != nil {
			return fmt.Errorf("delete email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["BulkQueuedEmailCount"] = len(rows)
			evt.Data["BulkQueuedEmailFilter"] = filters.AuditSummary()
		}
	}
	return nil
}

// AuditRecord summarises deleting all filtered queued emails.
func (BulkDeleteQueueTask) AuditRecord(data map[string]any) string {
	count, _ := data["BulkQueuedEmailCount"].(int)
	if filter, ok := data["BulkQueuedEmailFilter"].(string); ok && filter != "" {
		return fmt.Sprintf("deleted %d queued emails (%s)", count, filter)
	}
	if count > 0 {
		return fmt.Sprintf("deleted %d queued emails", count)
	}
	return "deleted queued emails"
}

func sqlNullInt32(value int) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(value), Valid: value != 0}
}
