package linker

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// RequireLinkViewOrReply ensures the user has view or reply permission for the link.
// It also sets the current link ID and thread/topic context.
func RequireLinkViewOrReply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		cd.LoadSelectionsFromRequest(r)

		link, id, err := cd.SelectedAdminLinkerItem(r)
		if err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}

		if link.DeletedAt.Valid {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}

		if !link.Listed.Valid {
			handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
			return
		}

		cd.SetCurrentThreadAndTopic(link.ThreadID, 0)

		canReply := cd.HasGrant("linker", "link", "reply", id)
		if !(cd.HasGrant("linker", "link", "view", id) ||
			canReply ||
			cd.SelectedThreadCanReply()) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
