package handlers

import (
	"net/http"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// GetPageSize returns the preferred page size from the request context.
func GetPageSize(r *http.Request) int {
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		return cd.PageSize()
	}
	return config.DefaultPageSize
}
