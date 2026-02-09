package user

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
var _ tasks.EmailTemplatesRequired = (*PermissionUpdateTask)(nil)

func (PermissionUpdateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	permid := r.PostFormValue("permid")
	role := r.PostFormValue("role")

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
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["targetUserID"] = infoID
				evt.Data["Username"] = username
				evt.Data["Role"] = role
			}
		} else {
			log.Printf("lookup role: %v", err2)
		}
	}

	return handlers.TemplateWithDataHandler("admin/runTaskPage.gohtml", data)
}

func roleInfoByPermID(ctx context.Context, q db.Querier, id int32) (int32, string, string, error) {
	row, err := q.GetPermissionByID(ctx, id)
	if err != nil {
		return 0, "", "", err
	}
	return row.UsersIdusers, row.Username.String, row.Name, nil
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

func (PermissionUpdateTask) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateUpdateUserRole.EmailTemplates(), true
}

func (PermissionUpdateTask) TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := NotificationTemplateUpdateUserRole.NotificationTemplate()
	return &v
}

func (PermissionUpdateTask) RequiredTemplates() []tasks.Template {
	return append(EmailTemplateUpdateUserRole.RequiredTemplates(), NotificationTemplateUpdateUserRole.RequiredTemplates()...)
}
