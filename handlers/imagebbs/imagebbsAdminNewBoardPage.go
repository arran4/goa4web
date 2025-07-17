package imagebbs

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func AdminNewBoardPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*handlers.CoreData
		Boards []*db.Imageboard
	}

	data := Data{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
	}
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

	data.Boards = boardRows

	handlers.TemplateHandler(w, r, "adminNewBoardPage.gohtml", data)
}

func AdminNewBoardMakePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	parentBoardId, _ := strconv.Atoi(r.PostFormValue("pbid"))

	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)

	err := queries.CreateImageBoard(r.Context(), db.CreateImageBoardParams{
		ImageboardIdimageboard: int32(parentBoardId),
		Title:                  sql.NullString{Valid: true, String: name},
		Description:            sql.NullString{Valid: true, String: desc},
	})
	if err != nil {
		log.Printf("Error: createImageBoard: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/admin/imagebbs/boards", http.StatusTemporaryRedirect)
}
