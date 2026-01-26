package user

import (
	"errors"
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// ErrMailNotConfigured is returned when test mail has no provider configured.
var ErrMailNotConfigured = errors.New("mail isn't configured")

// TestMailTask sends a test email to the current user.
type TestMailTask struct{ tasks.TaskString }

var testMailTask = &TestMailTask{TaskString: tasks.TaskString(TaskTestMail)}

var _ tasks.Task = (*TestMailTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*TestMailTask)(nil)
var _ tasks.EmailTemplatesRequired = (*TestMailTask)(nil)

func (TestMailTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	user, _ := cd.CurrentUser()
	if user == nil {
		return common.UserError{ErrorMessage: "email unknown"}
	}
	if cd.EmailProvider() == nil {
		return common.UserError{ErrorMessage: ErrMailNotConfigured.Error()}
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
}

func (TestMailTask) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateTest.EmailTemplates(), true
}

func (TestMailTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := EmailTemplateTest.NotificationTemplate()
	return &s
}

func (TestMailTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateTest.RequiredPages()
}
