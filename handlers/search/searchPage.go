package search

import (
	"net/http"

	handlers "github.com/arran4/goa4web/handlers"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*handlers.CoreData
		SearchWords string
	}

	data := Data{
		CoreData: r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
	}

	handlers.TemplateHandler(w, r, "searchPage", data)
}
