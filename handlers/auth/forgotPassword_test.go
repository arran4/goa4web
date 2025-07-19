package auth

import (
	"github.com/arran4/goa4web/internal/notifications"
	"testing"

	"github.com/arran4/goa4web/core/templates"
)

func requireEmailTemplates(t *testing.T, prefix string) {
	t.Helper()
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	textTmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})
	if htmlTmpls.Lookup(notifications.EmailTextTemplateFilenameGenerator(prefix)) == nil {
		t.Errorf("missing html template %s.gohtml", prefix)
	}
	if textTmpls.Lookup(notifications.EmailHTMLTemplateFilenameGenerator(prefix)) == nil {
		t.Errorf("missing text template %s.gotxt", prefix)
	}
	if textTmpls.Lookup(notifications.EmailSubjectTemplateFilenameGenerator(prefix)) == nil {
		t.Errorf("missing subject template %sSubject.gotxt", prefix)
	}
}

func TestForgotPasswordTemplatesExist(t *testing.T) {
	requireEmailTemplates(t, "passwordResetEmail")
	requireEmailTemplates(t, "adminNotificationUserRequestPasswordResetEmail")
}
