package admin

import (
	notif "github.com/arran4/goa4web/internal/notifications"
)

const (
	EmailTemplateAnnouncement        notif.EmailTemplateName        = "announcementEmail"
	NotificationTemplateAnnouncement notif.NotificationTemplateName = "announcement"

	EmailTemplateAdminAddIPBan notif.EmailTemplateName = "adminAddIPBanEmail"

	EmailTemplateAdminDeleteIPBan notif.EmailTemplateName = "adminRemoveIPBanEmail"
	EmailTemplateVerify           notif.EmailTemplateName = "verifyEmail"

	EmailTemplateAdminPasswordReset            notif.EmailTemplateName = "passwordResetEmail"
	EmailTemplateAdminUserRequestPasswordReset notif.EmailTemplateName = "adminNotificationUserRequestPasswordResetEmail"
)
