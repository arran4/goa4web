package search

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		SearchWords string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Search"

	handlers.TemplateHandler(w, r, "searchPage", Data{
		SearchWords: r.FormValue("searchwords"),
	})
}
