package handlers

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// SetPageTitle records a page-specific title used in templates.
func SetPageTitle(r *http.Request, prefix string) {
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
		cd.PageTitle = prefix
	}
}

// SetPageTitlef formats and records a page-specific title used in templates.
func SetPageTitlef(r *http.Request, format string, args ...any) {
	SetPageTitle(r, fmt.Sprintf(format, args...))
}
