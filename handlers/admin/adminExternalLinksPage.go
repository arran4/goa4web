package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type AdminExternalLinksPageTask struct{}

func (t *AdminExternalLinksPageTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		Links         []*db.ExternalLink
		Query         string
		ResultSummary string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "External Links"
	query := strings.TrimSpace(r.URL.Query().Get(externalLinksFilterQueryParam))
	queries := cd.Queries()
	rows, err := queries.AdminListExternalLinks(r.Context(), db.AdminListExternalLinksParams{Limit: 200, Offset: 0})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if query != "" {
		filtered := make([]*db.ExternalLink, 0, len(rows))
		queryLower := strings.ToLower(query)
		for _, link := range rows {
			if externalLinkMatchesQuery(link, queryLower) {
				filtered = append(filtered, link)
			}
		}
		rows = filtered
	}
	data := Data{
		Links: rows,
		Query: query,
	}
	action := r.URL.Query().Get(externalLinksActionQueryParam)
	successCount := queryIntValue(r, externalLinksSuccessQueryParam)
	failureCount := queryIntValue(r, externalLinksFailureQueryParam)
	if action != "" {
		actionLabel := action
		switch action {
		case externalLinksActionRefresh:
			actionLabel = "Refreshed"
		case externalLinksActionDelete:
			actionLabel = "Deleted"
		}
		data.ResultSummary = fmt.Sprintf("%s external links: %d succeeded, %d failed.", actionLabel, successCount, failureCount)
	}
	return AdminExternalLinksPageTmpl.Handler(data)
}

func (t *AdminExternalLinksPageTask) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "External Links", "/admin/external-links", &AdminPageTask{}
}

var _ tasks.Task = (*AdminExternalLinksPageTask)(nil)
var _ tasks.HasBreadcrumb = (*AdminExternalLinksPageTask)(nil)

type AdminExternalLinkDetailsPageBreadcrumb struct {
	LinkID int32
}

func (p *AdminExternalLinkDetailsPageBreadcrumb) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Link %d", p.LinkID), "", &AdminExternalLinksPageTask{}
}

const AdminExternalLinksPageTmpl tasks.Template = "admin/externalLinksPage.gohtml"

// externalLinksFilterQueryParam names the query parameter for URL search filtering.
const externalLinksFilterQueryParam = "q"

// externalLinksActionQueryParam names the query parameter for the bulk action type.
const externalLinksActionQueryParam = "action"

// externalLinksSuccessQueryParam names the query parameter for successful bulk actions.
const externalLinksSuccessQueryParam = "success"

// externalLinksFailureQueryParam names the query parameter for failed bulk actions.
const externalLinksFailureQueryParam = "failed"

// externalLinksActionRefresh labels the refresh bulk action.
const externalLinksActionRefresh = "refresh"

// externalLinksActionDelete labels the delete bulk action.
const externalLinksActionDelete = "delete"

func externalLinkMatchesQuery(link *db.ExternalLink, queryLower string) bool {
	if link == nil {
		return false
	}
	if strings.Contains(strings.ToLower(link.Url), queryLower) {
		return true
	}
	if strings.Contains(strconv.Itoa(int(link.ID)), queryLower) {
		return true
	}
	if link.CardTitle.Valid && strings.Contains(strings.ToLower(link.CardTitle.String), queryLower) {
		return true
	}
	if link.CardDescription.Valid && strings.Contains(strings.ToLower(link.CardDescription.String), queryLower) {
		return true
	}
	return false
}

func queryIntValue(r *http.Request, key string) int {
	value, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return 0
	}
	return value
}

type AdminExternalLinkDetailsPage struct {
	LinkID int32
	Data   any
}

func (p *AdminExternalLinkDetailsPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Link %d", p.LinkID), "", &AdminExternalLinksPageTask{}
}

func (p *AdminExternalLinkDetailsPage) PageTitle() string {
	return fmt.Sprintf("External Link %d", p.LinkID)
}

func (p *AdminExternalLinkDetailsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AdminExternalLinkDetailsPageTmpl.Handler(p.Data).ServeHTTP(w, r)
}

type AdminExternalLinkDetailsTask struct{}

func (t *AdminExternalLinkDetailsTask) Action(w http.ResponseWriter, r *http.Request) any {
	type Data struct {
		Link          *db.ExternalLink
		ResultSummary string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasAdminRole() {
		return handlers.ErrForbidden
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return handlers.ErrBadRequest
	}

	queries := cd.Queries()
	link, err := queries.GetExternalLinkByID(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return handlers.ErrNotFound
		}
		return fmt.Errorf("fetching external link: %w", err)
	}

	data := Data{
		Link: link,
	}

	// Result Summary Logic (similar to list page)
	action := r.URL.Query().Get(externalLinksActionQueryParam)
	successCount := queryIntValue(r, externalLinksSuccessQueryParam)
	failureCount := queryIntValue(r, externalLinksFailureQueryParam)
	if action != "" {
		actionLabel := action
		switch action {
		case externalLinksActionRefresh:
			actionLabel = "Refreshed"
		case externalLinksActionDelete:
			actionLabel = "Deleted"
		}
		data.ResultSummary = fmt.Sprintf("%s external link: %d succeeded, %d failed.", actionLabel, successCount, failureCount)
	}

	return &AdminExternalLinkDetailsPage{
		LinkID: int32(id),
		Data: data,
	}
}
