package forumcommon

import (
	"net/http"
	"strings"

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

func forumBasePath(cd *common.CoreData, r *http.Request) string {
	if cd != nil && cd.ForumBasePath != "" {
		return cd.ForumBasePath
	}
	if r != nil && strings.HasPrefix(r.URL.Path, "/private") {
		return "/private"
	}
	return "/forum"
}
