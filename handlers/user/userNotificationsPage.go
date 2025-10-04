package user

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

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
	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}
	data := struct{ Request *http.Request }{r}
	handlers.TemplateHandler(w, r, "notifications.gohtml", data)
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
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
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

func userNotificationEmailActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		handlers.RedirectToGet(w, r, "/usr/notifications")
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
	handlers.RedirectToGet(w, r, "/usr/notifications")
}
