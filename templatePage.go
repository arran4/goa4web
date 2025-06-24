package goa4web

import (
	"log"
	"net/http"
)

func templatePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if err := renderTemplate(w, r, "templatePage.gohtml", data); err != nil {
		log.Printf("template page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
