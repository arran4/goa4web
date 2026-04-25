package forum

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

func requireAdminEmailTemplates(t *testing.T, et *notif.EmailTemplates) {
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

func TestForumAdminTemplatesExist(t *testing.T) {
	admins := []tasks.Task{
		createThreadTask,
		replyTask,
		categoryChangeTask,
		categoryCreateTask,
		deleteCategoryTask,
		threadDeleteTask,
	}
	for _, p := range admins {
		et, _, ok := notif.AdminTemplates(p, eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
		if ok && et != nil {
			requireAdminEmailTemplates(t, et)
		}
	}
}
