package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RefreshExternalLinkTask clears cached preview fields so they will be reloaded.
type RefreshExternalLinkTask struct{ tasks.TaskString }

var refreshExternalLinkTask = &RefreshExternalLinkTask{TaskString: TaskUpdate}

var _ tasks.Task = (*RefreshExternalLinkTask)(nil)
var _ tasks.AuditableTask = (*RefreshExternalLinkTask)(nil)

func (RefreshExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { handlers.RenderErrorPage(w, r, handlers.ErrForbidden) })
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := cd.Queries()
	uid := sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0}
	successCount, failureCount := 0, 0
	for _, idStr := range r.Form["id"] {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			failureCount++
			continue
		}
		if err := queries.AdminClearExternalLinkCache(r.Context(), db.AdminClearExternalLinkCacheParams{UpdatedBy: uid, ID: int32(id)}); err != nil {
			failureCount++
			continue
		}
		successCount++
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["RefreshedID"] = appendID(evt.Data["RefreshedID"], id)
		}
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["RefreshSuccessCount"] = successCount
		evt.Data["RefreshFailureCount"] = failureCount
	}
	return handlers.RedirectHandler(externalLinksTaskRedirect(r, externalLinksActionRefresh, successCount, failureCount))
}

func (RefreshExternalLinkTask) AuditRecord(data map[string]any) string {
	successCount, failureCount := auditCounts(data, "RefreshSuccessCount", "RefreshFailureCount")
	if ids, ok := data["RefreshedID"].(string); ok {
		return fmt.Sprintf("refreshed external links %s (success %d, failed %d)", ids, successCount, failureCount)
	}
	return fmt.Sprintf("refreshed external links (success %d, failed %d)", successCount, failureCount)
}

// DeleteExternalLinkTask removes external link entries.
type DeleteExternalLinkTask struct{ tasks.TaskString }

var deleteExternalLinkTask = &DeleteExternalLinkTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteExternalLinkTask)(nil)
var _ tasks.AuditableTask = (*DeleteExternalLinkTask)(nil)

func (DeleteExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { handlers.RenderErrorPage(w, r, handlers.ErrForbidden) })
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := cd.Queries()
	successCount, failureCount := 0, 0
	for _, idStr := range r.Form["id"] {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			failureCount++
			continue
		}
		if err := queries.AdminDeleteExternalLink(r.Context(), int32(id)); err != nil {
			failureCount++
			continue
		}
		successCount++
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["DeletedID"] = appendID(evt.Data["DeletedID"], id)
		}
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["DeleteSuccessCount"] = successCount
		evt.Data["DeleteFailureCount"] = failureCount
	}
	return handlers.RedirectHandler(externalLinksTaskRedirect(r, externalLinksActionDelete, successCount, failureCount))
}

func (DeleteExternalLinkTask) AuditRecord(data map[string]any) string {
	successCount, failureCount := auditCounts(data, "DeleteSuccessCount", "DeleteFailureCount")
	if ids, ok := data["DeletedID"].(string); ok {
		return fmt.Sprintf("deleted external links %s (success %d, failed %d)", ids, successCount, failureCount)
	}
	return fmt.Sprintf("deleted external links (success %d, failed %d)", successCount, failureCount)
}

func externalLinksTaskRedirect(r *http.Request, action string, successCount, failureCount int) string {
	values := url.Values{}
	if filter := strings.TrimSpace(r.FormValue(externalLinksFilterQueryParam)); filter != "" {
		values.Set(externalLinksFilterQueryParam, filter)
	}
	values.Set(externalLinksActionQueryParam, action)
	values.Set(externalLinksSuccessQueryParam, strconv.Itoa(successCount))
	values.Set(externalLinksFailureQueryParam, strconv.Itoa(failureCount))
	return r.URL.Path + "?" + values.Encode()
}

func auditCounts(data map[string]any, successKey, failureKey string) (int, int) {
	if data == nil {
		return 0, 0
	}
	successCount, _ := data[successKey].(int)
	failureCount, _ := data[failureKey].(int)
	return successCount, failureCount
}
