package user

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type DismissTask struct{ tasks.TaskString }

var dismissTask = &DismissTask{TaskString: tasks.TaskString(TaskDismiss)}
var _ tasks.Task = (*DismissTask)(nil)

func userNotificationsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Notifications"
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	if cd.FeedsEnabled {
		cd.RSSFeedURL = cd.GenerateFeedURL("/usr/notifications/rss")
		cd.RSSFeedTitle = "Notifications RSS Feed"
		cd.AtomFeedURL = cd.GenerateFeedURL("/usr/notifications/atom")
		cd.AtomFeedTitle = "Notifications Atom Feed"
	}
	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	data := struct{ Request *http.Request }{r}
	handlers.TemplateHandler(w, r, "user/notifications.gohtml", data)
}

func (DismissTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return nil
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	n, err := queries.GetNotificationForLister(r.Context(), db.GetNotificationForListerParams{ID: int32(id), ListerID: uid})
	if err == nil && !n.ReadAt.Valid {
		if err := queries.SetNotificationReadForLister(r.Context(), db.SetNotificationReadForListerParams{ID: n.ID, ListerID: uid}); err != nil {
			log.Printf("mark notification read: %v", err)
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/notifications"}
}

func notificationsRssPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	var uid int32
	vars := mux.Vars(r)
	if username := vars["username"]; username != "" {
		user, err := handlers.VerifyFeedRequest(r, "/usr/notifications/rss")
		if err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		uid = user.Idusers
	} else {
		session, ok := core.GetSessionOrFail(w, r)
		if !ok {
			return
		}
		uid, _ = session.Values["UID"].(int32)
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	limit := int32(cd.Config.PageSizeDefault)
	notifs, err := queries.ListUnreadNotificationsForLister(r.Context(), db.ListUnreadNotificationsForListerParams{ListerID: uid, Limit: limit, Offset: 0})
	if err != nil {
		log.Printf("notify feed: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	feed := NotificationsFeed(r, notifs)
	if err := feed.WriteRss(w); err != nil {
		log.Printf("feed write: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}

func notificationsAtomPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	var uid int32
	vars := mux.Vars(r)
	if username := vars["username"]; username != "" {
		user, err := handlers.VerifyFeedRequest(r, "/usr/notifications/atom")
		if err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
			return
		}
		uid = user.Idusers
	} else {
		session, ok := core.GetSessionOrFail(w, r)
		if !ok {
			return
		}
		uid, _ = session.Values["UID"].(int32)
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	limit := int32(cd.Config.PageSizeDefault)
	notifs, err := queries.ListUnreadNotificationsForLister(r.Context(), db.ListUnreadNotificationsForListerParams{ListerID: uid, Limit: limit, Offset: 0})
	if err != nil {
		log.Printf("notify feed: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	feed := NotificationsFeed(r, notifs)
	if err := feed.WriteAtom(w); err != nil {
		log.Printf("feed write: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
}

func userNotificationOpenPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	n, err := queries.GetNotificationForLister(r.Context(), db.GetNotificationForListerParams{ID: int32(id), ListerID: uid})
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("notification open: %v", err)
		}
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	if !n.Link.Valid {
		http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
		return
	}
	data := struct {
		Request      *http.Request
		Notification *db.Notification
		RedirectURL  string
		TaskName     string
	}{
		Request:      r,
		Notification: n,
		RedirectURL:  n.Link.String,
		TaskName:     string(TaskDismiss),
	}
	handlers.TemplateHandler(w, r, "user/notificationOpen.gohtml", data)
}

func userNotificationEmailActionPage(w http.ResponseWriter, r *http.Request) {
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
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	val, _ := queries.GetMaxNotificationPriority(r.Context(), uid)
	var maxPr int32
	switch v := val.(type) {
	case int64:
		maxPr = int32(v)
	case int32:
		maxPr = v
	}
	if id != 0 {
		if err := queries.SetNotificationPriorityForLister(r.Context(), db.SetNotificationPriorityForListerParams{ListerID: uid, NotificationPriority: maxPr + 1, ID: int32(id)}); err != nil {
			log.Printf("set notification priority: %v", err)
		}
	}
	http.Redirect(w, r, "/usr/notifications", http.StatusSeeOther)
}
