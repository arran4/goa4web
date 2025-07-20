package forum

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
	nt := templates.GetCompiledNotificationTemplates(map[string]any{})
	if nt.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestForumTemplatesExist(t *testing.T) {
	providers := []notif.SubscribersNotificationTemplateProvider{
		createThreadTask,
		replyTask,
	}
	for _, p := range providers {
		checkEmailTemplates(t, p.SubscribedEmailTemplate())
		checkNotificationTemplate(t, p.SubscribedInternalNotificationTemplate())
	}
}
