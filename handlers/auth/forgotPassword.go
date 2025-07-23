package auth

import (
	"net/http"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func (EmailAssociationRequestTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationEmailAssociationRequestEmail")
}

func (EmailAssociationRequestTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationEmailAssociationRequestEmail")
	return &v
}

func (f ForgotPasswordTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationUserRequestPasswordResetEmail")
}

func (f ForgotPasswordTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationUserRequestPasswordResetEmail")
	return &v
}

func (f ForgotPasswordTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("passwordResetEmail")
}

func (f ForgotPasswordTask) SelfInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("password_reset")
	return &s
}

func (ForgotPasswordTask) SelfEmailBroadcast() bool { return true }

func (ForgotPasswordTask) Page(w http.ResponseWriter, r *http.Request) {
	handlers.TemplateHandler(w, r, "forgotPasswordPage.gohtml", r.Context().Value(consts.KeyCoreData))
}
