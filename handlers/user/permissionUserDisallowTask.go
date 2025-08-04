package user

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// PermissionUserDisallowTask removes a user's permission.
type PermissionUserDisallowTask struct{ tasks.TaskString }

var permissionUserDisallowTask = &PermissionUserDisallowTask{TaskString: TaskUserDisallow}

var _ tasks.Task = (*PermissionUserDisallowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*PermissionUserDisallowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*PermissionUserDisallowTask)(nil)

func (PermissionUserDisallowTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminPermissionDisallowEmail")
}

func (PermissionUserDisallowTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminPermissionDisallowEmail")
	return &v
}

func (PermissionUserDisallowTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	permid := r.PostFormValue("permid")
	id := cd.CurrentProfileUserID()
	back := "/admin/users/permissions"
	if id != 0 {
		back = fmt.Sprintf("/admin/user/%d/permissions", id)
	}
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     back,
	}
	if permidi, err := strconv.Atoi(permid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else {
		var (
			uname  string
			userID int32
			role   string
		)
		if rows, err := queries.GetUserRoles(r.Context()); err == nil {
			for _, row := range rows {
				if row.IduserRoles == int32(permidi) {
					role = row.Role
					userID = row.UsersIdusers
					if u, err := queries.SystemGetUserByID(r.Context(), row.UsersIdusers); err == nil && u.Username.Valid {
						uname = u.Username.String
					}
					break
				}
			}
		}
		if err := queries.AdminDeleteUserRole(r.Context(), int32(permidi)); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("CreateLanguage: %w", err).Error())
		} else if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["Username"] = uname
				evt.Data["Permission"] = role
				evt.Data["targetUserID"] = userID
				evt.Data["Role"] = role
			}
		} else {
			log.Printf("lookup role: %v", err)
		}
	}
	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}

func (PermissionUserDisallowTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (PermissionUserDisallowTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("deleteUserRoleEmail")
}

func (PermissionUserDisallowTask) TargetInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("delete_user_role")
	return &v
}
