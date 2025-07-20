package forum

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func requireEmailTemplatesFromProvider(t *testing.T, et *notif.EmailTemplates) {
	t.Helper()
	if et == nil {
		return
	}
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
	t.Helper()
	if name == nil {
		return
	}
	nt := templates.GetCompiledNotificationTemplates(map[string]any{})
	if nt.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestCreateThreadTemplatesExist(t *testing.T) {
	requireEmailTemplatesFromProvider(t, createThreadTask.SubscribedEmailTemplate())
	requireNotificationTemplate(t, createThreadTask.SubscribedInternalNotificationTemplate())
	requireEmailTemplatesFromProvider(t, createThreadTask.AdminEmailTemplate())
	requireNotificationTemplate(t, createThreadTask.AdminInternalNotificationTemplate())
}

func TestReplyTemplatesExist(t *testing.T) {
	requireEmailTemplatesFromProvider(t, replyTask.SubscribedEmailTemplate())
	requireNotificationTemplate(t, replyTask.SubscribedInternalNotificationTemplate())
	requireEmailTemplatesFromProvider(t, replyTask.AdminEmailTemplate())
	requireNotificationTemplate(t, replyTask.AdminInternalNotificationTemplate())
}

func TestRepliesMustAutoSubscribe(t *testing.T) {
	if _, ok := interface{}(replyTask).(notif.AutoSubscribeProvider); !ok {
		t.Fatalf("AutoSubscribeProvider must auto subscribe as users will want updates")
	}
}
