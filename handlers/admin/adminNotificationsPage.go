package admin

import (
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminNotificationsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Notifications []*db.Notification
		Total         int
		Unread        int
		Roles         []*db.Role
		Usernames     map[int32]string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Notifications"
	data := Data{}
	queries := cd.Queries()
	roles, err := cd.AllRoles()
	if err != nil {
		log.Printf("load roles: %v", err)
	}
	data.Roles = roles
	items, err := queries.AdminListRecentNotifications(r.Context(), 50)
	if err != nil {
		log.Printf("recent notifications: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	ids := make([]int32, 0, len(items))
	seen := map[int32]struct{}{}
	for _, n := range items {
		if _, ok := seen[n.UsersIdusers]; !ok {
			seen[n.UsersIdusers] = struct{}{}
			ids = append(ids, n.UsersIdusers)
		}
	}
	data.Usernames = map[int32]string{}
	if rows, err := queries.AdminListUsersByID(r.Context(), ids); err == nil {
		for _, r := range rows {
			if r.Username.Valid {
				data.Usernames[r.Idusers] = r.Username.String
			}
		}
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
	AdminNotificationsPageTmpl.Handle(w, r, data)
}

const AdminNotificationsPageTmpl tasks.Template = "admin/notificationsPage.gohtml"
