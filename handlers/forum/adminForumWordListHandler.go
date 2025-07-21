package forum

import (
	"database/sql"
	_ "embed"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

func AdminForumWordListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows []sql.NullString
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

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
