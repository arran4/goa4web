package user

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// ResendVerificationEmailTask resends the verification link for an unverified user email address.
type ResendVerificationEmailTask struct{ tasks.TaskString }

var resendVerificationEmailTask = &ResendVerificationEmailTask{TaskString: TaskResend}

var _ tasks.Task = (*ResendVerificationEmailTask)(nil)
var _ notif.DirectEmailNotificationTemplateProvider = (*ResendVerificationEmailTask)(nil)
var _ tasks.EmailTemplatesRequired = (*ResendVerificationEmailTask)(nil)

func (ResendVerificationEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	return addEmailTask.Resend(w, r)
}

func (ResendVerificationEmailTask) DirectEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateVerify.EmailTemplates(), true
}

func (ResendVerificationEmailTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateVerify.RequiredPages()
}

func (ResendVerificationEmailTask) DirectEmailAddress(evt eventbus.TaskEvent) (string, error) {
	if evt.Data != nil {
		if email, ok := evt.Data["email"].(string); ok {
			return email, nil
		}
	}
	return "", fmt.Errorf("email not provided")
}
