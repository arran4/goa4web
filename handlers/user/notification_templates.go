package user

import (
	notif "github.com/arran4/goa4web/internal/notifications"
)

const (
	EmailTemplateVerify                notif.EmailTemplateName        = "verifyEmail"
	EmailTemplateUpdateUserRole        notif.EmailTemplateName        = "updateUserRoleEmail"
	NotificationTemplateUpdateUserRole notif.NotificationTemplateName = "update_user_role"

	EmailTemplateAdminPermissionAllow notif.EmailTemplateName        = "adminPermissionAllowEmail"
	EmailTemplateSetUserRole          notif.EmailTemplateName        = "setUserRoleEmail"
	NotificationTemplateSetUserRole   notif.NotificationTemplateName = "set_user_role"

	EmailTemplateAdminPermissionDisallow notif.EmailTemplateName        = "adminPermissionDisallowEmail"
	EmailTemplateDeleteUserRole          notif.EmailTemplateName        = "deleteUserRoleEmail"
	NotificationTemplateDeleteUserRole   notif.NotificationTemplateName = "delete_user_role"

	EmailTemplateTest notif.EmailTemplateName = "testEmail"
)
