package user

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// PermissionUserAllowTask grants a user permission.
type PermissionUserAllowTask struct{ tasks.TaskString }

var permissionUserAllowTask = &PermissionUserAllowTask{TaskString: TaskUserAllow}

var _ tasks.Task = (*PermissionUserAllowTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*PermissionUserAllowTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*PermissionUserAllowTask)(nil)
var _ tasks.EmailTemplatesRequired = (*PermissionUserAllowTask)(nil)

func (PermissionUserAllowTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminPermissionAllow.EmailTemplates(), true
}

func (PermissionUserAllowTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminPermissionAllow.NotificationTemplate()
	return &v
}

func (PermissionUserAllowTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")
	cpu := cd.CurrentProfileUser()
	back := "/admin/users/permissions"
	if cpu.Idusers != 0 {
		back = fmt.Sprintf("/admin/user/%d/permissions", cpu.Idusers)
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["Username"] = username
		evt.Data["Permission"] = role
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: back,
	}
	if u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("SystemGetUserByUsername: %w", err).Error())
	} else if err := queries.SystemCreateUserRole(r.Context(), db.SystemCreateUserRoleParams{
		UsersIdusers: u.Idusers,
		Name:         role,
	}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("permissionUserAllow: %w", err).Error())
	} else if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["targetUserID"] = u.Idusers
		evt.Data["Username"] = u.Username.String
		evt.Data["Role"] = role
	}
	return handlers.TemplateWithDataHandler("admin/runTaskPage.gohtml", data)
}

func (PermissionUserAllowTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (PermissionUserAllowTask) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateSetUserRole.EmailTemplates(), true
}

func (PermissionUserAllowTask) TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := NotificationTemplateSetUserRole.NotificationTemplate()
	return &v
}

func (PermissionUserAllowTask) EmailTemplatesRequired() []tasks.Page {
	return append(EmailTemplateAdminPermissionAllow.RequiredPages(), append(EmailTemplateSetUserRole.RequiredPages(), NotificationTemplateSetUserRole.RequiredPages()...)...)
}
