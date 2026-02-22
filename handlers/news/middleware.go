package news

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// EnforceNewsPostAccess ensures the user has permission to edit the news post
// specified in the URL path. It sets the current news post ID context on success.
func EnforceNewsPostAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr, ok := vars["news"]
		if !ok {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}
		postID, err := strconv.Atoi(idStr)
		if err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}

		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok || cd == nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		if !cd.HasGrant("news", "post", "edit", int32(postID)) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		cd.SetCurrentNewsPost(int32(postID))
		next.ServeHTTP(w, r)
	})
}
