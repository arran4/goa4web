package imagebbs

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func requireEmailTemplates(t *testing.T, et *notif.EmailTemplates) {
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

func TestImageBbsTemplatesExist(t *testing.T) {
	requireEmailTemplates(t, notif.NewEmailTemplates("imageBoardNewEmail"))
	requireEmailTemplates(t, newBoardTask.AdminEmailTemplate())
}

func requireNotificationTemplate(t *testing.T, name string) {
	t.Helper()
	nt := templates.GetCompiledNotificationTemplates(map[string]any{})
	if nt.Lookup(name) == nil {
		t.Errorf("missing notification template %s", name)
	}
}
