package handlers

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// SetPageTitle prepends prefix to the global site title.
func SetPageTitle(r *http.Request, prefix string) {
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
		cd.Title = prefix + " - " + cd.Title
	}
}

// SetPageTitlef formats and prepends the prefix to the global site title.
func SetPageTitlef(r *http.Request, format string, args ...any) {
	SetPageTitle(r, fmt.Sprintf(format, args...))
}
