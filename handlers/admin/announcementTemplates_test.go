package admin

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/handlertest"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func checkEmailTemplates(t *testing.T, et *notif.EmailTemplates) {
	t.Helper()
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(handlertest.GetTemplateFuncs())
	textTmpls := templates.GetCompiledEmailTextTemplates(handlertest.GetTemplateFuncs())
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
	/*nt := templates.GetCompiledNotificationTemplates(map[string]any{
		"a4code2string": func(s string) string { return s },
		"truncateWords": func(i int, s string) string { return s },
		"cd":            func() any { return nil },
	})*/
	nt := templates.GetCompiledNotificationTemplates(handlertest.GetTemplateFuncs())
	if nt.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestAnnouncementTemplatesExist(t *testing.T) {
	admins := []notif.AdminEmailTemplateProvider{
		addAnnouncementTask,
		deleteAnnouncementTask,
	}
	for _, p := range admins {
		if et, _ := p.AdminEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); et != nil {
			checkEmailTemplates(t, et)
		}
		checkNotificationTemplate(t, p.AdminInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
	}
}
