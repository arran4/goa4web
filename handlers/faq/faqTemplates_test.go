package faq

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/handlertest"
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
	tmpl := templates.GetCompiledNotificationTemplates(handlertest.GetTemplateFuncs())
	if tmpl.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestAskTaskTemplatesCompile(t *testing.T) {
	var task AskTask
	if et, _ := task.AdminEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); et != nil {
		requireEmailTemplates(t, et)
	}
	requireNotificationTemplate(t, task.AdminInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
}

func TestAdminNotificationFaqAskEmailIncludesLink(t *testing.T) {
	url := "http://example.com/admin/faq/questions"
	data := notif.EmailData{URL: url, Item: map[string]any{"Question": "test?"}}
	textTmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(map[string]any{})

	var sb strings.Builder
	if err := textTmpls.ExecuteTemplate(&sb, "adminNotificationFaqAskEmail.gotxt", data); err != nil {
		t.Fatalf("render text: %v", err)
	}
	if !strings.Contains(sb.String(), url) {
		t.Errorf("text template missing url: %s", sb.String())
	}

	sb.Reset()
	if err := htmlTmpls.ExecuteTemplate(&sb, "adminNotificationFaqAskEmail.gohtml", data); err != nil {
		t.Fatalf("render html: %v", err)
	}
	if !strings.Contains(sb.String(), url) {
		t.Errorf("html template missing url: %s", sb.String())
	}
}
