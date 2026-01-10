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

// AdminBoardViewPage shows basic information for a board.
func AdminBoardViewPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Board *db.Imageboard
		Posts []*db.ListImagePostsByBoardForListerRow
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "View Image Board"

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

	rows, err := cd.Queries().ListImagePostsByBoardForLister(r.Context(), db.ListImagePostsByBoardForListerParams{
		ListerID:     cd.UserID,
		BoardID:      sql.NullInt32{Int32: board.Idimageboard, Valid: true},
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        5,
		Offset:       0,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	data := Data{Board: board, Posts: rows}
	ImageBBSAdminBoardViewPageTmpl.Handle(w, r, data)
}

const ImageBBSAdminBoardViewPageTmpl handlers.Page = "imagebbs/adminBoardViewPage.gohtml"
