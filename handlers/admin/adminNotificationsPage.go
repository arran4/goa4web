package admin

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

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
