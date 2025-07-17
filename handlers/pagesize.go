package handlers

import (
	"net/http"

	"github.com/arran4/goa4web/config"
	db "github.com/arran4/goa4web/internal/db"
)

// GetPageSize returns the preferred page size within configured bounds.
func GetPageSize(r *http.Request) int {
	size := config.AppRuntimeConfig.PageSizeDefault
	if pref, _ := r.Context().Value(ContextKey("preference")).(*db.Preference); pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < config.AppRuntimeConfig.PageSizeMin {
		size = config.AppRuntimeConfig.PageSizeMin
	}
	if size > config.AppRuntimeConfig.PageSizeMax {
		size = config.AppRuntimeConfig.PageSizeMax
	}
	return size
}
