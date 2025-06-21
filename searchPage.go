package main

import (
	"net/http"
)

func searchPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		SearchWords string
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	renderTemplate(w, r, "searchPage.gohtml", data)
}
