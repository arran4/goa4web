package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func BloggersBloggerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Rows []*db.ListBloggersForListerRow
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Bloggers"
	data := Data{}

	queries := cd.Queries()

	rows, err := queries.ListBloggersForLister(r.Context(), db.ListBloggersForListerParams{
		ListerID: cd.UserID,
		Limit:    1000,
		Offset:   0,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "bloggersBloggerPage.gohtml", data)
}
