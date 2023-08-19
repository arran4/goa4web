package main

import (
	"log"
	"net/http"
)

func userLogoutPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	// TODO cont.user.setUser(0);

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	// Custom Index???

	if err := getCompiledTemplates().ExecuteTemplate(w, "userLogoutPage.tmpl", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
