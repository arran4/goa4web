package search

import (
	"net/http"

	common "github.com/arran4/goa4web/handlers/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*hcommon.CoreData
		SearchWords string
	}

	data := Data{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
	}

	common.TemplateHandler(w, r, "searchPage", data)
}
