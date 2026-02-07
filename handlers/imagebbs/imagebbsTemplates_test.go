package imagebbs

import (
	"fmt"
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/handlertest"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func checkEmailTemplates(t *testing.T, et *notif.EmailTemplates) {
	t.Helper()
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(handlertest.GetTemplateFuncs(), templates.WithSilence(true))
	textTmpls := templates.GetCompiledEmailTextTemplates(handlertest.GetTemplateFuncs(), templates.WithSilence(true))
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
	tmpl := templates.GetCompiledNotificationTemplates(handlertest.GetTemplateFuncs(), templates.WithSilence(true))
	if tmpl.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestHappyPathImageBbsTemplatesExist(t *testing.T) {
	admins := []notif.AdminEmailTemplateProvider{
		newBoardTask,
		modifyBoardTask,
	}
	for i, p := range admins {
		t.Run(fmt.Sprintf("AdminProvider_%d", i), func(t *testing.T) {
			if et, _ := p.AdminEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); et != nil {
				checkEmailTemplates(t, et)
			}
			if p != newBoardTask {
				checkNotificationTemplate(t, p.AdminInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
			}
		})
	}

	selfProviders := []notif.SelfNotificationTemplateProvider{
		approvePostTask,
	}
	for i, p := range selfProviders {
		t.Run(fmt.Sprintf("SelfProvider_%d", i), func(t *testing.T) {
			if et, send := p.SelfEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); send {
				checkEmailTemplates(t, et)
			} else {
				t.Errorf("expected self email to be sent")
			}
			checkNotificationTemplate(t, p.SelfInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
		})
	}

	subs := []notif.SubscribersNotificationTemplateProvider{
		replyTask,
	}
	for i, p := range subs {
		t.Run(fmt.Sprintf("SubProvider_%d", i), func(t *testing.T) {
			if et, _ := p.SubscribedEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}); et != nil {
				checkEmailTemplates(t, et)
			}
			checkNotificationTemplate(t, p.SubscribedInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
		})
	}
}
