package blogs

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func BloggersBloggerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows []*db.BloggerCountRow
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	data.CoreData.PageTitle = "Bloggers"

	cd := data.CoreData
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	rows, err := queries.ListBloggers(r.Context(), db.ListBloggersParams{
		ViewerID: cd.UserID,
		Limit:    1000,
		Offset:   0,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "bloggersBloggerPage.gohtml", data)
}
