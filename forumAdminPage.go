package main

import (
	"net/http"
)

func forumAdminPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	CustomForumIndex(data.CoreData, r)

	renderTemplate(w, r, "forumAdminPage.gohtml", data)
}
