package imagebbs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminBoardsPage(w http.ResponseWriter, r *http.Request) {
	type BoardRow struct {
		*db.Imageboard
		Threads int32
	}
	type Data struct {
		Boards    []*BoardRow
		AllBoards []*db.Imageboard
		Page      int
		NextPage  int
		PrevPage  int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Image Boards"
	queries := cd.Queries()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const limit = 50
	boards, err := queries.AdminListBoards(r.Context(), db.AdminListBoardsParams{Limit: limit + 1, Offset: int32((page - 1) * limit)})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("listBoards error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	nextPage := 0
	if len(boards) > limit {
		nextPage = page + 1
		boards = boards[:limit]
	}
	data := Data{Page: page}
	for _, b := range boards {
		threads, err := queries.AdminCountThreadsByBoard(r.Context(), b.Idimageboard)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("countThreads error: %s", err)
			threads = 0
		}
		data.Boards = append(data.Boards, &BoardRow{Imageboard: b, Threads: int32(threads)})
	}
	if page > 1 {
		data.PrevPage = page - 1
	}
	if nextPage != 0 {
		data.NextPage = nextPage
	}

	allBoards, err := cd.ImageBoards()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("imageBoards error: %s", err)
	}
	data.AllBoards = allBoards

	handlers.TemplateHandler(w, r, "adminBoardsPage.gohtml", data)
}
