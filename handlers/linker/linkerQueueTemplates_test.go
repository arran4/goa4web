package linker

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/handlertest"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func checkQueueEmailTemplates(t *testing.T, prefix string) {
	t.Helper()
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(handlertest.GetTemplateFuncs(), templates.WithSilence(true))
	textTmpls := templates.GetCompiledEmailTextTemplates(handlertest.GetTemplateFuncs(), templates.WithSilence(true))
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

func TestLinkerQueueTemplatesExist(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		prefixes := []string{
			"linkerApprovedEmail",
			"linkerRejectedEmail",
			"adminNotificationLinkerApprovedEmail",
			"adminNotificationLinkerRejectedEmail",
		}
		for _, p := range prefixes {
			checkQueueEmailTemplates(t, p)
		}
	})
}
