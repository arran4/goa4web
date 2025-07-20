package forum

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

func TestForumTemplatesExist(t *testing.T) {
	tasks := []notif.SubscribersNotificationTemplateProvider{
		createThreadTask,
		replyTask,
	}
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	textTmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})
	for _, tp := range tasks {
		et := tp.SubscribedEmailTemplate()
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
}
