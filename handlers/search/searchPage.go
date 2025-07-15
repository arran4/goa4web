package search

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"net/http"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*hcommon.CoreData
		SearchWords string
	}

	data := Data{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
	}

	hcommon.TemplateHandler(w, r, "searchPage", data)
}
