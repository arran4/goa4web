package admin

import (
	"database/sql"
	_ "embed"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sections"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
	_ "github.com/arran4/goa4web/internal/dbdrivers" // Register SQL drivers.
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
		CoreData:   r.Context().Value(common.KeyCoreData).(*CoreData),
		AdminLinks: sections.AdminLinks(),
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
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

	err := templates.RenderTemplate(w, "adminPage", data, corecommon.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
