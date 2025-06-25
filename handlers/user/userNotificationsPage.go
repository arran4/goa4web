package user

import (
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	db "github.com/arran4/goa4web/internal/db"
)

func userNotificationsPage(w http.ResponseWriter, r *http.Request) {
	if !common.NotificationsEnabled() {
		http.NotFound(w, r)
		return
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	notifs, err := queries.GetUnreadNotifications(r.Context(), uid)
	if err != nil {
		log.Printf("get notifications: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		Notifications []*db.Notification
	}{
		CoreData:      r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Notifications: notifs,
	}
	if err := templates.RenderTemplate(w, "notifications.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func userNotificationsDismissActionPage(w http.ResponseWriter, r *http.Request) {
	if !common.NotificationsEnabled() {
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
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
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
	if !common.NotificationsEnabled() {
		http.NotFound(w, r)
		return
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
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
