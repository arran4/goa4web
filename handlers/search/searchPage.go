package search

import (
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		SearchWords string
	}

	data := Data{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData),
	}

	handlers.TemplateHandler(w, r, "searchPage", data)
}
