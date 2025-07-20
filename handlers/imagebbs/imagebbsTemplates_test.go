package imagebbs

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

func checkNotificationTemplate(t *testing.T, name *string) {
	if name == nil {
		return
	}
	tmpl := templates.GetCompiledNotificationTemplates(map[string]any{})
	if tmpl.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestImageBbsTemplatesExist(t *testing.T) {
	admins := []notif.AdminEmailTemplateProvider{
		newBoardTask,
		modifyBoardTask,
	}
	for _, p := range admins {
		checkEmailTemplates(t, p.AdminEmailTemplate())
		if p != newBoardTask {
			checkNotificationTemplate(t, p.AdminInternalNotificationTemplate())
		}
	}

	selfProviders := []notif.SelfNotificationTemplateProvider{
		approvePostTask,
	}
	for _, p := range selfProviders {
		checkEmailTemplates(t, p.SelfEmailTemplate())
		checkNotificationTemplate(t, p.SelfInternalNotificationTemplate())
	}

	subs := []notif.SubscribersNotificationTemplateProvider{
		replyTask,
	}
	for _, p := range subs {
		checkEmailTemplates(t, p.SubscribedEmailTemplate())
		checkNotificationTemplate(t, p.SubscribedInternalNotificationTemplate())
	}
}
