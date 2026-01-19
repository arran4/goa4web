package imagebbs

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/handlertest"
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
	tmpl := templates.GetCompiledNotificationTemplates(handlertest.GetTemplateFuncs())
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
		if et, _ := p.AdminEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); et != nil {
			checkEmailTemplates(t, et)
		}
		if p != newBoardTask {
			checkNotificationTemplate(t, p.AdminInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
		}
	}

	selfProviders := []notif.SelfNotificationTemplateProvider{
		approvePostTask,
	}
	for _, p := range selfProviders {
		if et, send := p.SelfEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); send {
			checkEmailTemplates(t, et)
		} else {
			t.Errorf("expected self email to be sent")
		}
		checkNotificationTemplate(t, p.SelfInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
	}

	subs := []notif.SubscribersNotificationTemplateProvider{
		replyTask,
	}
	for _, p := range subs {
		if et, _ := p.SubscribedEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); et != nil {
			checkEmailTemplates(t, et)
		}
		checkNotificationTemplate(t, p.SubscribedInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
	}
}
