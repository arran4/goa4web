package linker

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func requireEmailTemplates(t *testing.T, prefix string) {
	t.Helper()
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	textTmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})
	if htmlTmpls.Lookup(notif.EmailHTMLTemplateFilenameGenerator(prefix)) == nil {
		t.Errorf("missing html template %s.gohtml", prefix)
	}
	if textTmpls.Lookup(notif.EmailTextTemplateFilenameGenerator(prefix)) == nil {
		t.Errorf("missing text template %s.gotxt", prefix)
	}
	if textTmpls.Lookup(notif.EmailSubjectTemplateFilenameGenerator(prefix)) == nil {
		t.Errorf("missing subject template %sSubject.gotxt", prefix)
	}
}

func requireNotificationTemplate(t *testing.T, name *string) {
	if name == nil {
		return
	}
	nt := templates.GetCompiledNotificationTemplates(map[string]any{})
	if nt.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestLinkerTemplatesExist(t *testing.T) {
	requireEmailTemplates(t, "linkerAddEmail")
	requireNotificationTemplate(t, AddTask.SubscribedInternalNotificationTemplate())
	requireEmailTemplates(t, "adminNotificationLinkerAddEmail")
	requireNotificationTemplate(t, AddTask.AdminInternalNotificationTemplate())
}
