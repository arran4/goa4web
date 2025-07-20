package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"sort"
	"strconv"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

func adminUsersPermissionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows  []*db.GetPermissionsWithUsersRow
		Roles []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}

	rows, err := queries.GetPermissionsWithUsers(r.Context(), db.GetPermissionsWithUsersParams{Username: sql.NullString{}})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Username.String < rows[j].Username.String
	})
	data.Rows = rows
}

// PermissionUserAllowTask grants a user permission.
type PermissionUserAllowTask struct{ tasks.TaskString }

var permissionUserAllowTask = &PermissionUserAllowTask{TaskString: TaskUserAllow}

var _ tasks.Task = (*PermissionUserAllowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*PermissionUserAllowTask)(nil)

func (PermissionUserAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	username := r.PostFormValue("username")
	level := r.PostFormValue("role")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users/permissions",
	}
	if u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("GetUserByUsername: %w", err).Error())
	} else if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         level,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	} else if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["targetUserID"] = u.Idusers
			evt.Data["Username"] = u.Username.String
			evt.Data["Role"] = level
		}
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func (PermissionUserAllowTask) TargetUserIDs(evt eventbus.Event) []int32 {
	if id, ok := evt.Data["UserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["UserID"].(int); ok {
		return []int32{int32(id)}
	}
	if id, ok := evt.Data["UserID"].(float64); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (PermissionUserAllowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return nil
}

func (PermissionUserAllowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("permission_user_allow")
	return &v
}

// PermissionUserDisallowTask removes a user's permission.
type PermissionUserDisallowTask struct{ tasks.TaskString }

var permissionUserDisallowTask = &PermissionUserDisallowTask{TaskString: TaskUserDisallow}

var _ tasks.Task = (*PermissionUserDisallowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*PermissionUserDisallowTask)(nil)

func (PermissionUserDisallowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users/permissions",
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else {
		id, username, role, err2 := roleInfoByPermID(r.Context(), queries, int32(permidi))
		if err2 != nil {
			log.Printf("lookup role: %v", err2)
		}
		if err := queries.DeleteUserRole(r.Context(), int32(permidi)); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
		} else if err2 == nil {
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
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func (PermissionUserDisallowTask) TargetUserIDs(evt eventbus.Event) []int32 {
	if id, ok := evt.Data["UserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["UserID"].(int); ok {
		return []int32{int32(id)}
	}
	if id, ok := evt.Data["UserID"].(float64); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (PermissionUserDisallowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return nil
}

func (PermissionUserDisallowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("permission_user_disallow")
	return &v
}

// PermissionUpdateTask updates an existing permission entry.
type PermissionUpdateTask struct{ tasks.TaskString }

var permissionUpdateTask = &PermissionUpdateTask{TaskString: TaskUpdate}

var _ tasks.Task = (*PermissionUpdateTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*PermissionUpdateTask)(nil)

func (PermissionUpdateTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	permid := r.PostFormValue("permid")
	level := r.PostFormValue("role")

	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users/permissions",
	}

	if id, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else {
		infoID, username, _, err2 := roleInfoByPermID(r.Context(), queries, int32(id))
		if err := queries.UpdatePermission(r.Context(), db.UpdatePermissionParams{
			IduserRoles: int32(id),
			Name:        level,
		}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("UpdatePermission: %w", err).Error())
		} else if err2 == nil {
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
				if evt := cd.Event(); evt != nil {
					if evt.Data == nil {
						evt.Data = map[string]any{}
					}
					evt.Data["targetUserID"] = infoID
					evt.Data["Username"] = username
					evt.Data["Role"] = level
				}
			}
		} else {
			log.Printf("lookup role: %v", err2)
		}
	}

	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
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

func (PermissionUserAllowTask) TargetUserIDs(evt eventbus.Event) []int32 {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (PermissionUserAllowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("setUserRoleEmail")
}

func (PermissionUserAllowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("set_user_role")
	return &v
}

func (PermissionUserDisallowTask) TargetUserIDs(evt eventbus.Event) []int32 {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (PermissionUserDisallowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("deleteUserRoleEmail")
}

func (PermissionUserDisallowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("delete_user_role")
	return &v
}

func (PermissionUpdateTask) TargetUserIDs(evt eventbus.Event) []int32 {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}
	}
	return nil
}

func (PermissionUpdateTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("updateUserRoleEmail")
}

func (PermissionUpdateTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("update_user_role")
	return &v
}
