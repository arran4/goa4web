package search

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
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
		Stats Stats
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Search Admin"
	data := Data{}

	queries := cd.Queries()
	stats, err := queries.AdminGetSearchStats(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("database not available"))
		return
	}
	data.Stats.Words = stats.Words
	data.Stats.Comments = stats.Comments
	data.Stats.News = stats.News
	data.Stats.Blogs = stats.Blogs
	data.Stats.Linker = stats.Linker
	data.Stats.Writing = stats.Writings // maintain existing struct fields
	data.Stats.Writings = stats.Writings
	data.Stats.Images = stats.Images

	AdminSearchPageTmpl.Handle(w, r, data)
}

const AdminSearchPageTmpl handlers.Page = "admin/searchPage.gohtml"
