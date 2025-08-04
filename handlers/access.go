package handlers

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

// VerifyAccess wraps h and denies the request if the caller lacks any of
// the provided roles.
func VerifyAccess(h http.HandlerFunc, roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !common.Allowed(r, roles...) {
			RenderErrorPage(w, r, fmt.Errorf("Forbidden"))
			return
		}
		h(w, r)
	}
}
