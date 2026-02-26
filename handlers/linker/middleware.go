package linker

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func EnforceLinkerCommentsAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok {
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}

		// Ensure selections are loaded (populates currentLinkID from mux vars)
		cd.LoadSelectionsFromRequest(r)

		id, err := cd.SelectedAdminLinkerItemID(r)
		if err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
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
			log.Printf("GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser error: %v", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}

		// Setup context for SelectedThreadCanReply
		cd.SetCurrentSection("linker")

		canReply := cd.HasGrant("linker", "link", "reply", link.ID)

		if cd.HasGrant("linker", "link", "view", link.ID) ||
			canReply ||
			cd.SelectedThreadCanReply() {
			next.ServeHTTP(w, r)
			return
		}

		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
	})
}

func EnforceLinkViewAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok {
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}

		// Ensure selections are loaded (populates currentLinkID from mux vars)
		cd.LoadSelectionsFromRequest(r)

		id, err := cd.SelectedAdminLinkerItemID(r)
		if err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		if !cd.HasGrant("linker", "link", "view", id) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
