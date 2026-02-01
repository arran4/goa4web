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

// DeleteExternalLinkDetailTask removes an external link and redirects to the list page.
type DeleteExternalLinkDetailTask struct{ tasks.TaskString }

var deleteExternalLinkDetailTask = &DeleteExternalLinkDetailTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteExternalLinkDetailTask)(nil)
var _ tasks.AuditableTask = (*DeleteExternalLinkDetailTask)(nil)

func (DeleteExternalLinkDetailTask) Action(w http.ResponseWriter, r *http.Request) any {
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
	return handlers.RedirectHandler(externalLinksListRedirect(r, externalLinksActionDelete, successCount, failureCount))
}

func (DeleteExternalLinkDetailTask) AuditRecord(data map[string]any) string {
	successCount, failureCount := auditCounts(data, "DeleteSuccessCount", "DeleteFailureCount")
	if ids, ok := data["DeletedID"].(string); ok {
		return fmt.Sprintf("deleted external links %s (success %d, failed %d)", ids, successCount, failureCount)
	}
	return fmt.Sprintf("deleted external links (success %d, failed %d)", successCount, failureCount)
}

func externalLinksListRedirect(r *http.Request, action string, successCount, failureCount int) string {
	values := url.Values{}
	values.Set(externalLinksActionQueryParam, action)
	values.Set(externalLinksSuccessQueryParam, strconv.Itoa(successCount))
	values.Set(externalLinksFailureQueryParam, strconv.Itoa(failureCount))
	return "/admin/external-links?" + values.Encode()
}

// UpdateExternalLinkMetadataTask updates the metadata of an external link.
type UpdateExternalLinkMetadataTask struct{ tasks.TaskString }

var updateExternalLinkMetadataTask = &UpdateExternalLinkMetadataTask{TaskString: TaskSave}

var _ tasks.Task = (*UpdateExternalLinkMetadataTask)(nil)
var _ tasks.AuditableTask = (*UpdateExternalLinkMetadataTask)(nil)

func (UpdateExternalLinkMetadataTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { handlers.RenderErrorPage(w, r, handlers.ErrForbidden) })
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := cd.Queries()

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		return fmt.Errorf("invalid id: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	toNullString := func(s string) sql.NullString {
		return sql.NullString{String: s, Valid: s != ""}
	}

	arg := db.UpdateExternalLinkMetadataParams{
		ID:              int32(id),
		CardTitle:       toNullString(r.FormValue("card_title")),
		CardDescription: toNullString(r.FormValue("card_description")),
		CardImage:       toNullString(r.FormValue("card_image")),
		CardDuration:    toNullString(r.FormValue("card_duration")),
		CardUploadDate:  toNullString(r.FormValue("card_upload_date")),
		CardAuthor:      toNullString(r.FormValue("card_author")),
	}

	if err := queries.UpdateExternalLinkMetadata(r.Context(), arg); err != nil {
		return handlers.RedirectHandler(externalLinksDetailsRedirect(r, "save", 0, 1))
	}

	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["UpdatedID"] = id
		evt.Data["UpdateSuccessCount"] = 1
	}

	return handlers.RedirectHandler(externalLinksDetailsRedirect(r, "save", 1, 0))
}

func (UpdateExternalLinkMetadataTask) AuditRecord(data map[string]any) string {
	id, _ := data["UpdatedID"].(int)
	return fmt.Sprintf("updated external link metadata %d", id)
}

func externalLinksDetailsRedirect(r *http.Request, action string, successCount, failureCount int) string {
	values := url.Values{}
	values.Set(externalLinksActionQueryParam, action)
	values.Set(externalLinksSuccessQueryParam, strconv.Itoa(successCount))
	values.Set(externalLinksFailureQueryParam, strconv.Itoa(failureCount))
	return r.URL.Path + "?" + values.Encode()
}
