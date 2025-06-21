package main

import (
	"log"
	"net/http"
)

func adminNotificationsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Notifications []*Notification
	}
	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	items, err := queries.RecentNotifications(r.Context(), 50)
	if err != nil {
		log.Printf("recent notifications: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Notifications = items
	if err := renderTemplate(w, r, "adminNotificationsPage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
