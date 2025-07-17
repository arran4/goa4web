package search

import (
	"database/sql"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	"log"
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func adminSearchPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Words    int64
		WordList int64
		Comments int64
		News     int64
		Blogs    int64
		Linker   int64
		Writing  int64
		Writings int64
		Images   int64
	}

	type Data struct {
		*corecommon.CoreData
		Stats Stats
	}

	data := Data{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData),
	}

	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	ctx := r.Context()
	count := func(query string, dest *int64) {
		if err := queries.DB().QueryRowContext(ctx, query).Scan(dest); err != nil && err != sql.ErrNoRows {
			log.Printf("adminSearchPage count query error: %v", err)
		}
	}

	count("SELECT COUNT(*) FROM searchwordlist", &data.Stats.Words)
	count("SELECT COUNT(*) FROM comments_search", &data.Stats.Comments)
	count("SELECT COUNT(*) FROM site_news_search", &data.Stats.News)
	count("SELECT COUNT(*) FROM blogs_search", &data.Stats.Blogs)
	count("SELECT COUNT(*) FROM linker_search", &data.Stats.Linker)
	count("SELECT COUNT(*) FROM writing_search", &data.Stats.Writing)
	count("SELECT COUNT(*) FROM writing_search", &data.Stats.Writings)
	count("SELECT COUNT(*) FROM imagepost_search", &data.Stats.Images)

	handlers.TemplateHandler(w, r, "adminSearchPage", data)
}
