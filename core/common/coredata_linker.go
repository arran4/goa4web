package common

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/go-be-lazy"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// LinkerItemByID returns a linker item lazily loading it once per ID.
func (cd *CoreData) LinkerItemByID(id int32, ops ...lazy.Option[*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow]) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow, error) {
	fetch := func(i int32) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(cd.ctx, db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
			ViewerID:     cd.UserID,
			ID:           i,
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return row, nil
	}
	return lazy.Map(&cd.linkerItems, &cd.mapMu, id, fetch, ops...)
}

// CurrentLinkerItem returns the currently requested linker item lazily loaded.
func (cd *CoreData) CurrentLinkerItem(r *http.Request, ops ...lazy.Option[*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow]) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow, error) {
	if cd.currentLinkID == 0 {
		if r != nil {
			idStr := ""
			if vars := mux.Vars(r); vars != nil {
				idStr = vars["link"]
			}
			if idStr == "" {
				idStr = r.URL.Query().Get("link")
			}
			if idStr != "" {
				id, err := strconv.Atoi(idStr)
				if err != nil {
					return nil, fmt.Errorf("invalid link id: %w", err)
				}
				cd.currentLinkID = int32(id)
			}
		}
		if cd.currentLinkID == 0 {
			return nil, nil
		}
	}
	return cd.LinkerItemByID(cd.currentLinkID, ops...)
}

// CurrentLinkerItemLoaded returns the cached current linker item without database access.
func (cd *CoreData) CurrentLinkerItemLoaded() *db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow {
	if cd.linkerItems == nil {
		return nil
	}
	lv, ok := cd.linkerItems[cd.currentLinkID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}
