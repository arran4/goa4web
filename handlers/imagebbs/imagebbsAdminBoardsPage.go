package imagebbs

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminBoardsPage(w http.ResponseWriter, r *http.Request) {
	type BoardRow struct {
		*db.Imageboard
		Threads int32
		Visible bool
		Nsfw    bool
	}
	type Data struct {
		*common.CoreData
		Boards []*BoardRow
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	boardRows, err := data.CoreData.ImageBoards()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllImageBoards Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	for _, b := range boardRows {
		threads, err := queries.CountThreadsByBoard(r.Context(), b.Idimageboard)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("countThreads error: %s", err)
			threads = 0
		}
		data.Boards = append(data.Boards, &BoardRow{
			Imageboard: b,
			Threads:    int32(threads),
			Visible:    true,
			Nsfw:       false,
		})
	}

	handlers.TemplateHandler(w, r, "adminBoardsPage.gohtml", data)
}
