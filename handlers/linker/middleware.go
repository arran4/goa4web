package linker

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// RequireLinkAccess ensures the user has view permission for the link.
func RequireLinkAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		cd.LoadSelectionsFromRequest(r)
		link, err := cd.CurrentLinkerItem(r)
		if err != nil {
			log.Printf("RequireLinkAccess: error loading link: %v", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
		if link == nil {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}

		if !cd.HasGrant("linker", "link", "view", link.ID) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireLinkViewOrReply ensures the user has view or reply permission for the link.
func RequireLinkViewOrReply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		cd.LoadSelectionsFromRequest(r)
		link, err := cd.CurrentLinkerItem(r)
		if err != nil {
			log.Printf("RequireLinkViewOrReply: error loading link: %v", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
		if link == nil {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}

		cd.SetCurrentThreadAndTopic(link.ThreadID, 0)
		canReply := cd.HasGrant("linker", "link", "reply", link.ID)
		if !(cd.HasGrant("linker", "link", "view", link.ID) ||
			canReply ||
			cd.SelectedThreadCanReply()) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
