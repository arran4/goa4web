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
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	ctx := r.Context()
	count := func(query string, dest *int64) {
		if err := queries.DB().QueryRowContext(ctx, query).Scan(dest); err != nil && err != sql.ErrNoRows {
			log.Printf("adminPage count query error: %v", err)
		}
	}
	count("SELECT COUNT(*) FROM users", &data.Stats.Users)
	count("SELECT COUNT(*) FROM language", &data.Stats.Languages)
	// site_news renamed from siteNews in schema version 24
	count("SELECT COUNT(*) FROM site_news", &data.Stats.News)
	count("SELECT COUNT(*) FROM blogs", &data.Stats.Blogs)
	count("SELECT COUNT(*) FROM forumtopic", &data.Stats.ForumTopics)
	count("SELECT COUNT(*) FROM forumthread", &data.Stats.ForumThreads)
	count("SELECT COUNT(*) FROM writing", &data.Stats.Writings)

	handlers.TemplateHandler(w, r, "adminPage", data)
}
