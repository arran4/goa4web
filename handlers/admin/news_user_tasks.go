package admin

import (
	"context"
	"database/sql"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
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
var _ notifications.TargetUsersNotificationProvider = (*NewsUserAllowTask)(nil)

func (NewsUserAllowTask) AdminEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("newsPermissionEmail")
}

func (NewsUserAllowTask) AdminInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("news_permission")
	return &v
}

func (NewsUserAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")
	u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
	if err != nil {
		log.Printf("GetUserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         role,
	}); err != nil {
		log.Printf("permissionUserAllow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["targetUserID"] = u.Idusers
			evt.Data["Username"] = u.Username.String
			evt.Data["Role"] = role
		}
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
var _ notifications.TargetUsersNotificationProvider = (*NewsUserRemoveTask)(nil)

func (NewsUserRemoveTask) AdminEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("newsPermissionEmail")
}

func (NewsUserRemoveTask) AdminInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("news_permission")
	return &v
}

func (NewsUserRemoveTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	permid, err := strconv.Atoi(r.PostFormValue("permid"))
	if err != nil {
		log.Printf("strconv.Atoi(permid) Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	id, username, role, err := roleInfoByPermID(r.Context(), queries, int32(permid))
	if err != nil {
		log.Printf("lookup role: %v", err)
	}
	if err := queries.DeleteUserRole(r.Context(), int32(permid)); err != nil {
		log.Printf("permissionUserDisallow Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err == nil {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["targetUserID"] = id
				evt.Data["Username"] = username
				evt.Data["Role"] = role
			}
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

func roleInfoByPermID(ctx context.Context, q *db.Queries, id int32) (int32, string, string, error) {
	rows, err := q.GetPermissionsWithUsers(ctx, db.GetPermissionsWithUsersParams{Username: sql.NullString{}})
	if err != nil {
		return 0, "", "", err
	}
	for _, row := range rows {
		if row.IduserRoles == id {
			return row.UsersIdusers, row.Username.String, row.Name, nil
		}
	}
	return 0, "", "", sql.ErrNoRows
}

func (NewsUserAllowTask) TargetUserIDs(evt eventbus.TaskEvent) []int32 {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (NewsUserAllowTask) TargetEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("setUserRoleEmail")
}

func (NewsUserAllowTask) TargetInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("set_user_role")
	return &v
}

func (NewsUserRemoveTask) TargetUserIDs(evt eventbus.TaskEvent) []int32 {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (NewsUserRemoveTask) TargetEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("deleteUserRoleEmail")
}

func (NewsUserRemoveTask) TargetInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("delete_user_role")
	return &v
}
