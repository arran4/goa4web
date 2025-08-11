package notifications_test

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/writings"
)

// Ensure the shared reply templates are present so all implementations deliver notifications consistently.

func TestReplyTemplatesExist(t *testing.T) {
	var task writings.ReplyTask
	html := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	text := templates.GetCompiledEmailTextTemplates(map[string]any{})
	nt := templates.GetCompiledNotificationTemplates(map[string]any{})
	et, _ := task.SubscribedEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
	if html.Lookup(et.HTML) == nil {
		t.Errorf("missing html template %s", et.HTML)
	}
	if text.Lookup(et.Text) == nil {
		t.Errorf("missing text template %s", et.Text)
	}
	if text.Lookup(et.Subject) == nil {
		t.Errorf("missing subject template %s", et.Subject)
	}
	ntName := task.SubscribedInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
	if nt.Lookup(*ntName) == nil {
		t.Errorf("missing notification template %s", *ntName)
	}
}
