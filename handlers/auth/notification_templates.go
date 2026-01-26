package auth

import (
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

const (
	EmailTemplateAdminEmailAssociationRequest  notif.EmailTemplateName        = "adminNotificationEmailAssociationRequestEmail"
	EmailTemplateAdminUserRequestPasswordReset notif.EmailTemplateName        = "adminNotificationUserRequestPasswordResetEmail"
	EmailTemplatePasswordReset                 notif.EmailTemplateName        = "passwordResetEmail"
	NotificationTemplatePasswordReset          notif.NotificationTemplateName = "password_reset"
)

func (EmailAssociationRequestTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateAdminEmailAssociationRequest.RequiredTemplates()
}
