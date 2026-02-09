package imagebbs

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// imagebbsApproveAction is the grant action needed for image board approvals and administration.
const imagebbsApproveAction = "approve"

func CheckBoardAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd == nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		boardID, err := imageBoardIDFromRequest(r, cd)
		if err != nil || boardID == 0 {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}

		if !cd.HasGrant("imagebbs", "board", "view", boardID) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireImagebbsGrant(action string) mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd == nil {
			return false
		}
		boardID, err := imageBoardIDFromRequest(r, cd)
		if err != nil {
			return false
		}
		return cd.HasGrant("imagebbs", "board", action, boardID)
	}
}

func imageBoardIDFromRequest(r *http.Request, cd *common.CoreData) (int32, error) {
	queries := cd.Queries()
	if queries == nil {
		return 0, fmt.Errorf("queries not available for image board lookup")
	}
	vars := mux.Vars(r)
	if boardStr, ok := vars["board"]; ok && boardStr != "" {
		boardID, err := strconv.Atoi(boardStr)
		if err != nil {
			return 0, fmt.Errorf("invalid board id %q: %w", boardStr, err)
		}
		return int32(boardID), nil
	}
	if boardStr, ok := vars["boardno"]; ok && boardStr != "" {
		boardID, err := strconv.Atoi(boardStr)
		if err != nil {
			return 0, fmt.Errorf("invalid boardno id %q: %w", boardStr, err)
		}
		return int32(boardID), nil
	}
	if postStr, ok := vars["post"]; ok && postStr != "" {
		postID, err := strconv.Atoi(postStr)
		if err != nil {
			return 0, fmt.Errorf("invalid post id %q: %w", postStr, err)
		}
		post, err := queries.AdminGetImagePost(r.Context(), int32(postID))
		if err != nil {
			return 0, fmt.Errorf("fetch image post %d: %w", postID, err)
		}
		if !post.ImageboardIdimageboard.Valid {
			return 0, fmt.Errorf("image post %d missing board id", postID)
		}
		return post.ImageboardIdimageboard.Int32, nil
	}
	return 0, nil
}
