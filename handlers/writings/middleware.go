package writings

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// EnforceArticleViewAccess ensures the user has permission to view the article
// and updates the thread context.
func EnforceArticleViewAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok || cd == nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		cd.LoadSelectionsFromRequest(r)
		writing, err := cd.Article()
		if err != nil || writing == nil {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}

		cd.SetCurrentThreadAndTopic(writing.ForumthreadID, 0)
		if !(cd.HasGrant("writing", "article", "view", writing.Idwriting) || cd.SelectedThreadCanReply()) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
