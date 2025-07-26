package handlers

import (
	"net/http"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// GetPageSize returns the preferred page size from the current CoreData or the default.
func GetPageSize(r *http.Request) int {
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
		return cd.PageSize()
	}
	return config.DefaultPageSize
}
