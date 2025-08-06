package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminBoardDashboardPage shows board info with recent posts.
func AdminBoardDashboardPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Board *db.Imageboard
		Posts []*db.ListImagePostsByBoardForListerRow
	}

	vars := mux.Vars(r)
	bidStr := vars["board"]
	bid, _ := strconv.Atoi(bidStr)
	if bid == 0 {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Image Board"
	queries := cd.Queries()

	board, err := queries.GetImageBoardById(r.Context(), int32(bid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		log.Printf("get image board: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	posts, err := queries.ListImagePostsByBoardForLister(r.Context(), db.ListImagePostsByBoardForListerParams{
		ListerID:     cd.UserID,
		BoardID:      int32(bid),
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        5,
		Offset:       0,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("list posts: %v", err)
	}

	data := Data{Board: board, Posts: posts}
	handlers.TemplateHandler(w, r, "adminBoardDashboardPage.gohtml", data)
}
