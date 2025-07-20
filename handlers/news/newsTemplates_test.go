package news

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func checkEmailTemplates(t *testing.T, et *notif.EmailTemplates) {
	t.Helper()
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	textTmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})
	if htmlTmpls.Lookup(et.HTML) == nil {
		t.Errorf("missing html template %s", et.HTML)
	}
	if textTmpls.Lookup(et.Text) == nil {
		t.Errorf("missing text template %s", et.Text)
	}
	if textTmpls.Lookup(et.Subject) == nil {
		t.Errorf("missing subject template %s", et.Subject)
	}
}

func TestNewsTemplatesExist(t *testing.T) {
	subs := []notif.SubscribersNotificationTemplateProvider{
		newPostTask,
		replyTask,
	}
	for _, p := range subs {
		checkEmailTemplates(t, p.SubscribedEmailTemplate())
	}

	admins := []notif.AdminEmailTemplateProvider{
		newPostTask,
		editTask,
		replyTask,
		editReplyTask,
		cancelTask,
		userAllowTask,
		userDisallowTask,
		announcementAddTask,
		announcementDeleteTask,
	}
	for _, p := range admins {
		checkEmailTemplates(t, p.AdminEmailTemplate())
	}
}
