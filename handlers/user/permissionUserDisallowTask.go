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
var _ tasks.EmailTemplatesRequired = (*PermissionUserDisallowTask)(nil)

func (PermissionUserDisallowTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminPermissionDisallow.EmailTemplates(), true
}

func (PermissionUserDisallowTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminPermissionDisallow.NotificationTemplate()
	return &v
}

func (PermissionUserDisallowTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	permid := r.PostFormValue("permid")
	cpu := cd.CurrentProfileUser()
	back := "/admin/users/permissions"
	if cpu.Idusers != 0 {
		back = fmt.Sprintf("/admin/user/%d/permissions", cpu.Idusers)
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: back,
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
		} else if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Username"] = uname
			evt.Data["Permission"] = role
			evt.Data["targetUserID"] = userID
			evt.Data["Role"] = role
		} else {
			log.Printf("lookup role: %v", err)
		}
	}
	return handlers.TemplateWithDataHandler("admin/runTaskPage.gohtml", data)
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

func (PermissionUserDisallowTask) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateDeleteUserRole.EmailTemplates(), true
}

func (PermissionUserDisallowTask) TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := NotificationTemplateDeleteUserRole.NotificationTemplate()
	return &v
}

func (PermissionUserDisallowTask) EmailTemplatesRequired() []tasks.Page {
	return append(EmailTemplateAdminPermissionDisallow.RequiredPages(), append(EmailTemplateDeleteUserRole.RequiredPages(), NotificationTemplateDeleteUserRole.RequiredPages()...)...)
}
