package main

import (
	"net/http"
)

func templatePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	renderTemplate(w, r, "templatePage.gohtml", data)
}
