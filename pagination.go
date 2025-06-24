package goa4web

import "net/http"

// getPageSize returns the preferred page size within configured bounds.
func getPageSize(r *http.Request) int {
	size := appRuntimeConfig.PageSizeDefault
	if pref, _ := r.Context().Value(ContextValues("preference")).(*Preference); pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < appRuntimeConfig.PageSizeMin {
		size = appRuntimeConfig.PageSizeMin
	}
	if size > appRuntimeConfig.PageSizeMax {
		size = appRuntimeConfig.PageSizeMax
	}
	return size
}
