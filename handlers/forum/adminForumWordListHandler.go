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

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Word List"
	data := Data{
		CoreData: cd,
	}

	queries := cd.Queries()

	rows, err := queries.AdminCompleteWordList(r.Context())
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
