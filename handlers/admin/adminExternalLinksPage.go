package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminExternalLinksPage lists cached external links.
func AdminExternalLinksPage(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("list external links: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
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
	AdminExternalLinksPageTmpl.Handle(w, r, data)
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
