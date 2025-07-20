package admin

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"strconv"
)

// newsUserAllowTask grants a role to a user and notifies admins.
type newsUserAllowTask struct{ tasks.TaskString }

var _ tasks.Task = (*newsUserAllowTask)(nil)
var _ notifications.AdminEmailTemplateProvider = (*newsUserAllowTask)(nil)

func (newsUserAllowTask) AdminEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("newsPermissionEmail")
}

func (newsUserAllowTask) AdminInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("news_permission")
	return &v
}

func (newsUserAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		log.Printf("GetUserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         level,
	}); err != nil {
		log.Printf("permissionUserAllow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

// newsUserRemoveTask revokes a role from a user and notifies admins.
type newsUserRemoveTask struct{ tasks.TaskString }

var _ tasks.Task = (*newsUserRemoveTask)(nil)
var _ notifications.AdminEmailTemplateProvider = (*newsUserRemoveTask)(nil)

func (newsUserRemoveTask) AdminEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("newsPermissionEmail")
}

func (newsUserRemoveTask) AdminInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("news_permission")
	return &v
}

func (newsUserRemoveTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	permid, err := strconv.Atoi(r.PostFormValue("permid"))
	if err != nil {
		log.Printf("strconv.Atoi(permid) Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := queries.DeleteUserRole(r.Context(), int32(permid)); err != nil {
		log.Printf("permissionUserDisallow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}
