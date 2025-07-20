package news_test

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	admin "github.com/arran4/goa4web/handlers/admin"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func checkEmailTemplates(t *testing.T, et *notif.EmailTemplates) {
	t.Helper()
	html := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	text := templates.GetCompiledEmailTextTemplates(map[string]any{})
	if html.Lookup(et.HTML) == nil {
		t.Errorf("missing html template %s", et.HTML)
	}
	if text.Lookup(et.Text) == nil {
		t.Errorf("missing text template %s", et.Text)
	}
	if text.Lookup(et.Subject) == nil {
		t.Errorf("missing subject template %s", et.Subject)
	}
}

func checkNotificationTemplate(t *testing.T, name *string) {
	if name == nil {
		return
	}
	nt := templates.GetCompiledNotificationTemplates(map[string]any{})
	if nt.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestNewsUserLevelTasksTemplates(t *testing.T) {
	allow := admin.NewsUserAllowTask{TaskString: admin.TaskNewsUserAllow}
	checkEmailTemplates(t, allow.AdminEmailTemplate())
	checkNotificationTemplate(t, allow.AdminInternalNotificationTemplate())

	remove := admin.NewsUserRemoveTask{TaskString: admin.TaskNewsUserRemove}
	checkEmailTemplates(t, remove.AdminEmailTemplate())
	checkNotificationTemplate(t, remove.AdminInternalNotificationTemplate())
}
