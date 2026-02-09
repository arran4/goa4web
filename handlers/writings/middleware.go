package writings

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// RequireWritingView ensures the requester may view the writing referenced in the URL.
func RequireWritingView(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok {
			http.NotFound(w, r)
			return
		}
		cd.LoadSelectionsFromRequest(r)
		writing, err := cd.CurrentWriting()
		if err != nil {
			log.Printf("RequireWritingView load writing: %v", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
		if writing == nil {
			log.Printf("RequireWritingView: no writing found")
			handlers.RenderErrorPage(w, r, fmt.Errorf("No writing found"))
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

// RequireWritingAuthor ensures the requester may edit the writing referenced in the URL.
func RequireWritingAuthor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok {
			http.NotFound(w, r)
			return
		}
		cd.LoadSelectionsFromRequest(r)
		writing, err := cd.CurrentWriting()
		if err != nil {
			log.Printf("RequireWritingAuthor load writing: %v", err)
			http.NotFound(w, r)
			return
		}
		if writing == nil {
			http.NotFound(w, r)
			return
		}
		if !cd.HasGrant("writing", "article", "edit", writing.Idwriting) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
