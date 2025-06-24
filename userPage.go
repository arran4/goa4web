package goa4web

import (
	"log"
	"net/http"
)

func userPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if data.CoreData.UserID == 0 {
		session, err := GetSession(r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		redirectToLogin(w, r, session)
		return
	}

	if err := renderTemplate(w, r, "userPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
