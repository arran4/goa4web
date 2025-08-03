package admin

import (
	"database/sql"
	_ "embed"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Users        int64
		Languages    int64
		News         int64
		Blogs        int64
		ForumTopics  int64
		ForumThreads int64
		Writings     int64
	}

	type Data struct {
		*common.CoreData
		AdminSections []common.AdminSection
		Stats         Stats
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin"
	data := Data{
		CoreData:      cd,
		AdminSections: cd.Nav.AdminSections(),
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).CustomQueries()
	ctx := r.Context()
	count := func(table string, dest *int64) {
		if c, err := queries.AdminCountTable(ctx, table); err == nil {
			*dest = c
		} else if err != sql.ErrNoRows {
			log.Printf("adminPage count %s error: %v", table, err)
		}
	}
	count("users", &data.Stats.Users)
	count("language", &data.Stats.Languages)
	count("site_news", &data.Stats.News)
	count("blogs", &data.Stats.Blogs)
	count("forumtopic", &data.Stats.ForumTopics)
	count("forumthread", &data.Stats.ForumThreads)
	count("writing", &data.Stats.Writings)

	handlers.TemplateHandler(w, r, "adminPage", data)
}
