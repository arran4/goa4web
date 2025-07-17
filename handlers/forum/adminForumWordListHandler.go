package forum

import (
	"database/sql"
	_ "embed"
	"errors"
	"net/http"

	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func AdminForumWordListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows []sql.NullString
	}

	data := Data{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*CoreData),
	}

	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

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

	common.TemplateHandler(w, r, "forumWordListPage.gohtml", data)
}
