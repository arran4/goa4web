package main

import (
	"log"
	"net/http"
)

func adminEmailQueuePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Emails []*PendingEmail
	}
	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	items, err := queries.ListUnsentPendingEmails(r.Context())
	if err != nil {
		log.Printf("list pending emails: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Emails = items
	if err := renderTemplate(w, r, "adminEmailQueuePage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
