package handlers

import (
	"net/http"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// GetPageSize returns the preferred page size within configured bounds.
func GetPageSize(r *http.Request) int {
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
		return cd.PageSize()
	}
	cfg := config.AppRuntimeConfig
	size := cfg.PageSizeDefault
	if size < cfg.PageSizeMin {
		size = cfg.PageSizeMin
	}
	if size > cfg.PageSizeMax {
		size = cfg.PageSizeMax
	}
	return size
}
