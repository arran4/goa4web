package imagebbs

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminBoardListPage lists images for a specific board.
func AdminBoardListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Board    *db.Imageboard
		Posts    []*db.ListImagePostsByBoardForListerRow
		Page     int
		NextPage int
		PrevPage int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["board"])
	if bid == 0 {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}

	boards, err := cd.ImageBoards()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	var board *db.Imageboard
	for _, b := range boards {
		if int(b.Idimageboard) == bid {
			board = b
			break
		}
	}
	if board == nil {
		http.NotFound(w, r)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const limit = 50
	rows, err := cd.Queries().ListImagePostsByBoardForLister(r.Context(), db.ListImagePostsByBoardForListerParams{
		ListerID:     cd.UserID,
		BoardID:      sql.NullInt32{Int32: board.Idimageboard, Valid: true},
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        limit + 1,
		Offset:       int32((page - 1) * limit),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	nextPage := 0
	if len(rows) > limit {
		nextPage = page + 1
		rows = rows[:limit]
	}
	data := Data{
		Board: board,
		Posts: rows,
		Page:  page,
	}
	if page > 1 {
		data.PrevPage = page - 1
	}
	if nextPage != 0 {
		data.NextPage = nextPage
	}
	cd.PageTitle = "Board Images"
	ImageBBSAdminBoardListPageTmpl.Handle(w, r, data)
}

const ImageBBSAdminBoardListPageTmpl handlers.Page = "imagebbs/adminBoardListPage.gohtml"
