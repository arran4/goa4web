package main

import "net/http"

// getPageSize returns the preferred page size within configured bounds.
func getPageSize(r *http.Request) int {
	size := appPaginationConfig.Default
	if pref, _ := r.Context().Value(ContextValues("preference")).(*Preference); pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < appPaginationConfig.Min {
		size = appPaginationConfig.Min
	}
	if size > appPaginationConfig.Max {
		size = appPaginationConfig.Max
	}
	return size
}
