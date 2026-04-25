package forum

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
	funcs := map[string]any{
		"truncateWords": func(i int, s string) string { return s },
	}
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(funcs, templates.WithSilence(true))
	textTmpls := templates.GetCompiledEmailTextTemplates(funcs, templates.WithSilence(true))
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

func TestForumTemplatesExist(t *testing.T) {
	providers := []tasks.Task{
		createThreadTask,
		replyTask,
	}
	for _, p := range providers {
		et, nt, ok := notif.SubscriberTemplates(p, eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
		if ok && et != nil {
			checkEmailTemplates(t, et)
		}
		checkNotificationTemplate(t, nt)
	}
}
