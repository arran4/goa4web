package forum

import (
	"database/sql"
	_ "embed"
	"errors"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func AdminForumWordListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows []sql.NullString
	}

	data := Data{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*CoreData),
	}

	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)

	rows, err := queries.CompleteWordList(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "forumWordListPage.gohtml", data)
}
