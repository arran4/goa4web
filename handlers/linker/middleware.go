package linker

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

type contextKey string

const keyLink contextKey = "linker_link"

func EnforceLinkerCommentsAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok {
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}

		// Ensure selections are loaded (populates currentLinkID from mux vars)
		cd.LoadSelectionsFromRequest(r)

		// Wait, I found 'SelectedLinkID' in CoreData? No, I checked before and it wasn't there.
		// I must check cd.LoadSelectionsFromRequest logic.
		// It sets private field 'currentLinkID'.
		// But 'SelectedAdminLinkerItemID' exists.
		// 'SelectedLinkerItemsForCurrentUser' exists.
		// There is no public getter for 'currentLinkID' other than indirectly via SelectedLinkerCategory etc.

		// But Wait, I need the ID to query.
		// I can get it from mux vars again to be safe.
		// Or I can assume LoadSelectionsFromRequest worked.

		// I'll extract it manually from Mux vars or use cd.SelectedAdminLinkerItemID(r) if it works for non-admin?
		// SelectedAdminLinkerItemID checks vars["link"].

		id, err := cd.SelectedAdminLinkerItemID(r)
		if err != nil {
             // Fallback to mux vars directly if helper fails or is admin specific (name implies admin but logic seems generic)
             // Logic:
             // if v, ok := mux.Vars(r)["link"]; ok { ... }
             // else if ... form ...

             // If error, it means no ID.
             handlers.RenderErrorPage(w, r, handlers.ErrForbidden) // Or 404
             return
		}

		queries := cd.Queries()

		// Use ForUser query to match existing behavior (enforces view grant implicitly)
		link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
			ViewerID:     cd.UserID,
			ID:           id,
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
				return
			}
			log.Printf("GetLinkerItemById error: %v", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}

		// Setup context for SelectedThreadCanReply
		cd.SetCurrentSection("linker")
		cd.SetCurrentThreadAndTopic(link.ThreadID, 0)
		// currentLinkID is set by LoadSelectionsFromRequest call above.

		canReply := cd.HasGrant("linker", "link", "reply", link.ID)

		if cd.HasGrant("linker", "link", "view", link.ID) ||
			canReply ||
			cd.SelectedThreadCanReply() {

			ctx := context.WithValue(r.Context(), keyLink, link)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
	})
}
