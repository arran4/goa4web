package goa4web

import (
	"database/sql"
	_ "embed"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
)

func adminPage(w http.ResponseWriter, r *http.Request) {
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
		Stats Stats
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
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

	err := templates.RenderTemplate(w, "page.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
