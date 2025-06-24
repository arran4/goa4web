package goa4web

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func templatePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if err := templates.RenderTemplate(w, "templatePage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("template page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
