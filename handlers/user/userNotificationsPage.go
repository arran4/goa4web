package user

import (
	"log"
	"net/http"
	"strconv"

	handlers "github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type NotificationDismissTask struct{ tasks.TaskString }
type NotificationEmailTask struct{ tasks.TaskString }

var (
	notificationDismissTask = &NotificationDismissTask{TaskString: TaskDismiss}
	notificationEmailTask   = &NotificationEmailTask{TaskString: TaskSaveAll}
)

func userNotificationsPage(w http.ResponseWriter, r *http.Request) {
	if !handlers.NotificationsEnabled() {
		http.NotFound(w, r)
		return
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	notifs, err := queries.GetUnreadNotifications(r.Context(), uid)
	if err != nil {
		log.Printf("get notifications: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	emails, _ := queries.GetUserEmailsByUserID(r.Context(), uid)
	var maxPr int32
	for _, e := range emails {
		if e.NotificationPriority > maxPr {
			maxPr = e.NotificationPriority
		}
	}
	data := struct {
		*handlers.CoreData
		Notifications []*db.Notification
		Emails        []*db.UserEmail
		MaxPriority   int32
	}{
		CoreData:      r.Context().Value(handlers.KeyCoreData).(*handlers.CoreData),
		Notifications: notifs,
		Emails:        emails,
		MaxPriority:   maxPr,
	}
	handlers.TemplateHandler(w, r, "notifications.gohtml", data)
}

func (NotificationDismissTask) Action(w http.ResponseWriter, r *http.Request) {
	if !handlers.NotificationsEnabled() {
		http.NotFound(w, r)
		return
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	n, err := queries.GetUnreadNotifications(r.Context(), uid)
	if err == nil {
		for _, no := range n {
			if int(no.ID) == id {
				_ = queries.MarkNotificationRead(r.Context(), no.ID)
				break
			}
		}
	}
	http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
}

func notificationsRssPage(w http.ResponseWriter, r *http.Request) {
	if !handlers.NotificationsEnabled() {
		http.NotFound(w, r)
		return
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
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

func (NotificationEmailTask) Action(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	idStr := r.FormValue("email_id")
	id, _ := strconv.Atoi(idStr)
	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)
	val, _ := queries.GetMaxNotificationPriority(r.Context(), uid)
	var maxPr int32
	switch v := val.(type) {
	case int64:
		maxPr = int32(v)
	case int32:
		maxPr = v
	}
	if id != 0 {
		_ = queries.SetNotificationPriority(r.Context(), db.SetNotificationPriorityParams{NotificationPriority: maxPr + 1, ID: int32(id)})
	}
	http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
}
