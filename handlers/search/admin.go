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

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).CustomQueries()
	ctx := r.Context()
	count := func(table string, dest *int64) {
		if c, err := queries.AdminCountTable(ctx, table); err == nil {
			*dest = c
		} else if err != sql.ErrNoRows {
			log.Printf("adminSearchPage count %s error: %v", table, err)
		}
	}

	count("searchwordlist", &data.Stats.Words)
	count("comments_search", &data.Stats.Comments)
	count("site_news_search", &data.Stats.News)
	count("blogs_search", &data.Stats.Blogs)
	count("linker_search", &data.Stats.Linker)
	count("writing_search", &data.Stats.Writing)
	count("writing_search", &data.Stats.Writings)
	count("imagepost_search", &data.Stats.Images)

	handlers.TemplateHandler(w, r, "adminSearchPage", data)
}
