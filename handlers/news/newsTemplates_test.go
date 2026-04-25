package news

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/handlertest"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
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
	nt := templates.GetCompiledNotificationTemplates(handlertest.GetTemplateFuncs(), templates.WithSilence(true))
	if nt.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestHappyPathNewsTemplatesExist(t *testing.T) {
	subs := []tasks.Task{
		newPostTask,
		replyTask,
	}
	for _, p := range subs {
		et, nt, ok := notif.SubscriberTemplates(p, eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
		if ok && et != nil {
			checkEmailTemplates(t, et)
		}
		checkNotificationTemplate(t, nt)
	}

	admins := []tasks.Task{
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
		et, nt, ok := notif.AdminTemplates(p, eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
		if ok && et != nil {
			checkEmailTemplates(t, et)
		}
		checkNotificationTemplate(t, nt)
	}
}
