package goa4web

import (
	"log"
	"net/http"
)

func userLogoutPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("logout request")
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	session, err := GetSession(r)
	if err != nil {
		sessionError(w, r, err)
	}
	delete(session.Values, "UID")
	delete(session.Values, "LoginTime")
	delete(session.Values, "ExpiryTime")
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	if session.ID != "" {
		_ = queries.DeleteSessionByID(r.Context(), session.ID)
	}

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("logout success")

	data.CoreData.UserID = 0
	data.CoreData.SecurityLevel = ""

	if err := renderTemplate(w, r, "logoutPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
