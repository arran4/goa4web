package admin

import (
	"database/sql"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"strconv"
)

// NewsUserAllowTask grants a role to a user and notifies admins.
type NewsUserAllowTask struct{ tasks.TaskString }

// TaskNewsUserAllow identifies a request to grant a user a role.
const TaskNewsUserAllow tasks.TaskString = "allow"

var newsUserAllow = &NewsUserAllowTask{TaskString: TaskNewsUserAllow}

var _ tasks.Task = (*NewsUserAllowTask)(nil)
var _ notifications.AdminEmailTemplateProvider = (*NewsUserAllowTask)(nil)

func (NewsUserAllowTask) AdminEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("newsPermissionEmail")
}

func (NewsUserAllowTask) AdminInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("news_permission")
	return &v
}

func (NewsUserAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	username := r.PostFormValue("username")
	levelID := r.PostFormValue("role")
	level, err := cd.ResolveRoleName(levelID)
	if err != nil {
		log.Printf("resolve role %s: %v", levelID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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

// NewsUserRemoveTask revokes a role from a user and notifies admins.
type NewsUserRemoveTask struct{ tasks.TaskString }

// TaskNewsUserRemove identifies a request to revoke a user's role.
const TaskNewsUserRemove tasks.TaskString = "remove"

var newsUserRemove = &NewsUserRemoveTask{TaskString: TaskNewsUserRemove}

var _ tasks.Task = (*NewsUserRemoveTask)(nil)
var _ notifications.AdminEmailTemplateProvider = (*NewsUserRemoveTask)(nil)

func (NewsUserRemoveTask) AdminEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("newsPermissionEmail")
}

func (NewsUserRemoveTask) AdminInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("news_permission")
	return &v
}

func (NewsUserRemoveTask) Action(w http.ResponseWriter, r *http.Request) {
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
