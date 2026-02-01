package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

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
