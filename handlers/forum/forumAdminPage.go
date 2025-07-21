package forum

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Categories int64
		Topics     int64
		Threads    int64
	}

	type Data struct {
		*common.CoreData
		Stats Stats
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	ctx := r.Context()
	count := func(q string, dest *int64) {
		if err := queries.DB().QueryRowContext(ctx, q).Scan(dest); err != nil && err != sql.ErrNoRows {
			log.Printf("forumAdminPage count query error: %v", err)
		}
	}

	count("SELECT COUNT(*) FROM forumcategory", &data.Stats.Categories)
	count("SELECT COUNT(*) FROM forumtopic", &data.Stats.Topics)
	count("SELECT COUNT(*) FROM forumthread", &data.Stats.Threads)

	handlers.TemplateHandler(w, r, "forumAdminPage", data)
}
