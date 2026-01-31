package auth

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/handlertest"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func requireEmailTemplates(t *testing.T, et *notif.EmailTemplates) {
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

func requireNotificationTemplate(t *testing.T, name *string) {
	if name == nil {
		return
	}
	nt := templates.GetCompiledNotificationTemplates(handlertest.GetTemplateFuncs())
	if nt.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestForgotPasswordTemplatesExist(t *testing.T) {
	admins := []notif.AdminEmailTemplateProvider{
		forgotPasswordTask,
		emailAssociationRequestTask,
	}
	for _, p := range admins {
		if et, _ := p.AdminEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); et != nil {
			requireEmailTemplates(t, et)
		}
		requireNotificationTemplate(t, p.AdminInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
	}

	selfProviders := []notif.SelfNotificationTemplateProvider{
		forgotPasswordTask,
	}
	for _, p := range selfProviders {
		if et, send := p.SelfEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); send {
			requireEmailTemplates(t, et)
		} else {
			t.Errorf("expected self email to be sent")
		}
		requireNotificationTemplate(t, p.SelfInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
	}
}
