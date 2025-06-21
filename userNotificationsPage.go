package main

import (
	"log"
	"net/http"
	"strconv"
)

func userNotificationsPage(w http.ResponseWriter, r *http.Request) {
	if !notificationsEnabled() {
		http.NotFound(w, r)
		return
	}
	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	notifs, err := queries.GetUnreadNotifications(r.Context(), uid)
	if err != nil {
		log.Printf("get notifications: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*CoreData
		Notifications []*Notification
	}{
		CoreData:      r.Context().Value(ContextValues("coreData")).(*CoreData),
		Notifications: notifs,
	}
	if err := renderTemplate(w, r, "userNotifications.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func userNotificationsDismissActionPage(w http.ResponseWriter, r *http.Request) {
	if !notificationsEnabled() {
		http.NotFound(w, r)
		return
	}
	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	n, err := queries.GetUnreadNotifications(r.Context(), uid)
	if err == nil {
		for _, no := range n {
			if int(no.ID) == id {
				_ = queries.MarkNotificationRead(r.Context(), no.ID)
				break
			}
		}
	}
	http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
}

func notificationsRssPage(w http.ResponseWriter, r *http.Request) {
	if !notificationsEnabled() {
		http.NotFound(w, r)
		return
	}
	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	notifs, err := queries.GetUnreadNotifications(r.Context(), uid)
	if err != nil {
		log.Printf("notify feed: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	feed := notificationsFeed(r, notifs)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
