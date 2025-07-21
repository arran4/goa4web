package admin

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"
	"strings"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

// MarkReadTask marks notifications as read.
type MarkReadTask struct{ tasks.TaskString }

var markReadTask = &MarkReadTask{TaskString: TaskDismiss}

// ensures MarkReadTask implements the tasks.Task interface
var _ tasks.Task = (*MarkReadTask)(nil)

// PurgeNotificationsTask removes old read notifications.
type PurgeNotificationsTask struct{ tasks.TaskString }

var purgeNotificationsTask = &PurgeNotificationsTask{TaskString: TaskPurge}

// ensures PurgeNotificationsTask implements the tasks.Task interface
var _ tasks.Task = (*PurgeNotificationsTask)(nil)

// SendNotificationTask creates a site notification for users.
type SendNotificationTask struct{ tasks.TaskString }

var sendNotificationTask = &SendNotificationTask{TaskString: TaskNotify}

// ensures SendNotificationTask implements the tasks.Task interface
var _ tasks.Task = (*SendNotificationTask)(nil)

func AdminNotificationsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Notifications []*db.Notification
		Total         int
		Unread        int
		Roles         []*db.Role
	}
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	roles, err := data.AllRoles()
	if err != nil {
		log.Printf("load roles: %v", err)
	}
	data.Roles = roles
	items, err := queries.RecentNotifications(r.Context(), 50)
	if err != nil {
		log.Printf("recent notifications: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	unread := 0
	for _, n := range items {
		if !n.ReadAt.Valid {
			unread++
		}
	}
	data.Notifications = items
	data.Total = len(items)
	data.Unread = unread
	handlers.TemplateHandler(w, r, "notificationsPage.gohtml", data)
}

func (MarkReadTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.MarkNotificationRead(r.Context(), int32(id)); err != nil {
			log.Printf("mark read: %v", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (PurgeNotificationsTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := queries.PurgeReadNotifications(r.Context()); err != nil {
		log.Printf("purge notifications: %v", err)
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (SendNotificationTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	message := r.PostFormValue("message")
	link := r.PostFormValue("link")
	role := r.PostFormValue("role")
	names := r.PostFormValue("users")

	var ids []int32
	if names != "" {
		for _, name := range strings.Split(names, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			u, err := queries.GetUserByUsername(r.Context(), sql.NullString{String: name, Valid: true})
			if err != nil {
				log.Printf("get user %s: %v", name, err)
				continue
			}
			ids = append(ids, u.Idusers)
		}
	} else if role != "" && role != "anonymous" {
		rows, err := queries.ListUserIDsByRole(r.Context(), role)
		if err != nil {
			log.Printf("list role: %v", err)
		} else {
			ids = append(ids, rows...)
		}
	} else {
		rows, err := queries.AllUserIDs(r.Context())
		if err != nil {
			log.Printf("list users: %v", err)
		} else {
			ids = append(ids, rows...)
		}
	}
	for _, id := range ids {
		err := queries.InsertNotification(r.Context(), db.InsertNotificationParams{
			UsersIdusers: id,
			Link:         sql.NullString{String: link, Valid: link != ""},
			Message:      sql.NullString{String: message, Valid: message != ""},
		})
		if err != nil {
			log.Printf("insert notification: %v", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}
