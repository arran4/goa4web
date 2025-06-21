package main

import (
	"net/http"
)

func imagebbsAdminPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomImageBBSIndex(data.CoreData, r)

	renderTemplate(w, r, "imagebbsAdminPage.gohtml", data)
}
