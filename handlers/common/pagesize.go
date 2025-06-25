package common

import (
	"net/http"

	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
)

// GetPageSize returns the preferred page size within configured bounds.
func GetPageSize(r *http.Request) int {
	size := runtimeconfig.AppRuntimeConfig.PageSizeDefault
	if pref, _ := r.Context().Value(ContextKey("preference")).(*db.Preference); pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < runtimeconfig.AppRuntimeConfig.PageSizeMin {
		size = runtimeconfig.AppRuntimeConfig.PageSizeMin
	}
	if size > runtimeconfig.AppRuntimeConfig.PageSizeMax {
		size = runtimeconfig.AppRuntimeConfig.PageSizeMax
	}
	return size
}
