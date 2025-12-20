package forum

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

func AdminForumWordListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Rows []sql.NullString
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Word List"
	data := Data{}

	queries := cd.Queries()

	rows, err := queries.AdminCompleteWordList(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			w.WriteHeader(http.StatusInternalServerError)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "admin/forumWordListPage.gohtml", data)
}
