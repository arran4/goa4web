package admin

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/notifications"
)

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

func (NewsUserAllowTask) AdminEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("newsPermissionEmail")
}

func (NewsUserAllowTask) AdminInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("news_permission")
	return &v
}

func (NewsUserRemoveTask) AdminEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("newsPermissionEmail")
}

func (NewsUserRemoveTask) AdminInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("news_permission")
	return &v
}

func (NewsUserAllowTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (NewsUserAllowTask) TargetEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("setUserRoleEmail")
}

func (NewsUserAllowTask) TargetInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("set_user_role")
	return &v
}

func (NewsUserRemoveTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (NewsUserRemoveTask) TargetEmailTemplate() *notifications.EmailTemplates {
	return notifications.NewEmailTemplates("deleteUserRoleEmail")
}

func (NewsUserRemoveTask) TargetInternalNotificationTemplate() *string {
	v := notifications.NotificationTemplateFilenameGenerator("delete_user_role")
	return &v
}

// AuditRecord summarises granting a role to a user.
func (NewsUserAllowTask) AuditRecord(data map[string]any) string {
	u, _ := data["Username"].(string)
	role, _ := data["Role"].(string)
	if u != "" && role != "" {
		return fmt.Sprintf("granted %s to %s", role, u)
	}
	return "granted user role"
}

// AuditRecord summarises revoking a user role.
func (NewsUserRemoveTask) AuditRecord(data map[string]any) string {
	u, _ := data["Username"].(string)
	role, _ := data["Role"].(string)
	if u != "" && role != "" {
		return fmt.Sprintf("revoked %s from %s", role, u)
	}
	return "revoked user role"
}
