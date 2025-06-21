package main

import (
	"net/http"
)

func userPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	renderTemplate(w, r, "userPage.gohtml", data)
}
