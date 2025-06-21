package main

import (
	"log"
	"net/http"
)

func adminSiteSettingsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		feedsEnabled = r.PostFormValue("feeds_enabled") != ""
		http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
		return
	}

	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	data.CoreData.FeedsEnabled = feedsEnabled

	if err := renderTemplate(w, r, "adminSiteSettingsPage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
