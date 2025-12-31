package handlers

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

func RenderPermissionDenied(w http.ResponseWriter, r *http.Request) {
	RenderErrorPage(w, r, WrapForbidden(ErrLoginRequired))
}

// VerifyAccess wraps h and denies the request if the caller lacks any of
// the provided roles. err is shown on the rendered error page; if err is nil
// a generic "Forbidden" message is displayed.
func VerifyAccess(h http.HandlerFunc, err error, roles ...string) http.HandlerFunc {
	if err == nil {
		err = ErrForbidden
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if !common.Allowed(r, roles...) {
			w.WriteHeader(http.StatusForbidden)
			RenderErrorPage(w, r, err)
			return
		}
		h(w, r)
	}
}
