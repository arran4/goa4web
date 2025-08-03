package search

import (
	"github.com/arran4/goa4web/core/consts"
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
	counts, err := queries.AdminSearchIndexCounts(r.Context())
	if err != nil {
		http.Error(w, "database not available", http.StatusInternalServerError)
		return
	}
	data.Stats.Words = counts.Words
	data.Stats.WordList = int64(counts.WordList)
	data.Stats.Comments = counts.Comments
	data.Stats.News = counts.News
	data.Stats.Blogs = counts.Blogs
	data.Stats.Linker = counts.Linker
	data.Stats.Writing = counts.Writing
	data.Stats.Writings = counts.Writings
	data.Stats.Images = counts.Images

	handlers.TemplateHandler(w, r, "adminSearchPage", data)
}
