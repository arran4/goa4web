package linker

import (
	notif "github.com/arran4/goa4web/internal/notifications"
)

const (
	EmailTemplateLinkerApproved                  notif.EmailTemplateName        = "linkerApprovedEmail"
	NotificationTemplateLinkerApproved           notif.NotificationTemplateName = "linker_approved"
	EmailTemplateAdminNotificationLinkerApproved notif.EmailTemplateName        = "adminNotificationLinkerApprovedEmail"

	EmailTemplateLinkerAdminAdd        notif.EmailTemplateName        = "linkerAdminAddEmail"
	NotificationTemplateLinkerAdminAdd notif.NotificationTemplateName = "linker_admin_add"

	EmailTemplateLinkerDeleted        notif.EmailTemplateName        = "linkerDeletedEmail"
	NotificationTemplateLinkerDeleted notif.NotificationTemplateName = "linker_deleted"

	EmailTemplateLinkerRejected                  notif.EmailTemplateName        = "linkerRejectedEmail"
	NotificationTemplateLinkerRejected           notif.NotificationTemplateName = "linker_rejected"
	EmailTemplateAdminNotificationLinkerRejected notif.EmailTemplateName        = "adminNotificationLinkerRejectedEmail"
)
