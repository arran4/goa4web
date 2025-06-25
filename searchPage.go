package goa4web

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
)

func searchPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		SearchWords string
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if err := templates.RenderTemplate(w, "searchPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
