package search

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
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
		*common.CoreData
		Stats Stats
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	data.CoreData.PageTitle = "Search Admin"

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
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
