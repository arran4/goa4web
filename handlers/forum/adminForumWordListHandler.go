package forum

import (
	"database/sql"
	_ "embed"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
)

func AdminForumWordListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows []sql.NullString
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

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

	err = templates.RenderTemplate(w, "forumWordListPage.gohtml", data, corecommon.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
