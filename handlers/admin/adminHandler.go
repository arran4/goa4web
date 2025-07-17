package admin

import (
	"database/sql"
	_ "embed"
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	nav "github.com/arran4/goa4web/internal/navigation"
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
		*CoreData
		AdminLinks []corecommon.IndexItem
		Stats      Stats
	}

	data := Data{
		CoreData:   r.Context().Value(corecommon.KeyCoreData).(*CoreData),
		AdminLinks: nav.AdminLinks(),
	}
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	ctx := r.Context()
	count := func(query string, dest *int64) {
		if err := queries.DB().QueryRowContext(ctx, query).Scan(dest); err != nil && err != sql.ErrNoRows {
			log.Printf("adminPage count query error: %v", err)
		}
	}
	count("SELECT COUNT(*) FROM users", &data.Stats.Users)
	count("SELECT COUNT(*) FROM language", &data.Stats.Languages)
	count("SELECT COUNT(*) FROM siteNews", &data.Stats.News)
	count("SELECT COUNT(*) FROM blogs", &data.Stats.Blogs)
	count("SELECT COUNT(*) FROM forumtopic", &data.Stats.ForumTopics)
	count("SELECT COUNT(*) FROM forumthread", &data.Stats.ForumThreads)
	count("SELECT COUNT(*) FROM writing", &data.Stats.Writings)

	common.TemplateHandler(w, r, "adminPage", data)
}
