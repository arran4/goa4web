package faq

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func requireEmailTemplates(t *testing.T, et *notif.EmailTemplates) {
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

func requireNotificationTemplate(t *testing.T, name *string) {
	if name == nil {
		return
	}
	tmpl := templates.GetCompiledNotificationTemplates(map[string]any{})
	if tmpl.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestAnswerTaskTemplatesCompile(t *testing.T) {
	var task AnswerTask
	requireEmailTemplates(t, task.AdminEmailTemplate())
	requireNotificationTemplate(t, task.AdminInternalNotificationTemplate())
	requireEmailTemplates(t, task.SelfEmailTemplate())
	requireNotificationTemplate(t, task.SelfInternalNotificationTemplate())
}

func TestAskTaskTemplatesCompile(t *testing.T) {
	var task AskTask
	requireEmailTemplates(t, task.AdminEmailTemplate())
	requireNotificationTemplate(t, task.AdminInternalNotificationTemplate())
}
