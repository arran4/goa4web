package writings

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/handlertest"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func TestReplyTemplatesCompile(t *testing.T) {
	// Ensure the ReplyTask exposes templates that actually exist so users
	// receive notification emails when someone responds.
	et, _ := replyTask.SubscribedEmailTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	textTmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})
	if htmlTmpls.Lookup(et.HTML) == nil {
		t.Fatalf("missing html template %s", et.HTML)
	}
	if textTmpls.Lookup(et.Text) == nil {
		t.Fatalf("missing text template %s", et.Text)
	}
	if textTmpls.Lookup(et.Subject) == nil {
		t.Fatalf("missing subject template %s", et.Subject)
	}

	nt := templates.GetCompiledNotificationTemplates(handlertest.GetTemplateFuncs())
	it := replyTask.SubscribedInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess})
	if nt.Lookup(*it) == nil {
		t.Fatalf("missing notification template %s", *it)
	}
}

func requireAutoSubscribeProvider(t *testing.T, task any) {
	t.Helper()
	if _, ok := task.(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("%T should auto subscribe so participants stay updated", task)
	}
}

func TestReplyTemplatesAutoSubscribe(t *testing.T) {
	requireAutoSubscribeProvider(t, replyTask)
}
