package forumcommon

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// BasePathMiddleware sets the forum base path on CoreData for downstream handlers.
func BasePathMiddleware(base string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
				cd.ForumBasePath = base
			}
			next.ServeHTTP(w, r)
		})
	}
}
