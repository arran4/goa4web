package user

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// PermissionUpdateTask updates an existing permission entry.
type PermissionUpdateTask struct{ tasks.TaskString }

var permissionUpdateTask = &PermissionUpdateTask{TaskString: TaskUpdate}

var _ tasks.Task = (*PermissionUpdateTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*PermissionUpdateTask)(nil)

func (PermissionUpdateTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	permid := r.PostFormValue("permid")
	role := r.PostFormValue("role")

	idStr := mux.Vars(r)["user"]
	back := "/admin/users/permissions"
	if idStr != "" {
		back = "/admin/user/" + idStr + "/permissions"
	}
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     back,
	}

	if id, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else {
		infoID, username, _, err2 := roleInfoByPermID(r.Context(), queries, int32(id))
		if err := queries.AdminUpdateUserRole(r.Context(), db.AdminUpdateUserRoleParams{
			IduserRoles: int32(id),
			Name:        role,
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
					evt.Data["Role"] = role
				}
			}
		} else {
			log.Printf("lookup role: %v", err2)
		}
	}

	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}

func roleInfoByPermID(ctx context.Context, q db.Querier, id int32) (int32, string, string, error) {
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

func (PermissionUpdateTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (PermissionUpdateTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("updateUserRoleEmail")
}

func (PermissionUpdateTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("update_user_role")
	return &v
}
