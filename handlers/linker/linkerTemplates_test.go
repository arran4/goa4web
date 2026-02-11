package linker

import (
	"testing"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/handlertest"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func requireEmailTemplates(t *testing.T, prefix string) {
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

func requireNotificationTemplate(t *testing.T, name *string) {
	if name == nil {
		return
	}
	nt := templates.GetCompiledNotificationTemplates(handlertest.GetTemplateFuncs(), templates.WithSilence(true))
	if nt.Lookup(*name) == nil {
		t.Errorf("missing notification template %s", *name)
	}
}

func TestLinkerTemplatesExist(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		requireEmailTemplates(t, "linkerAddEmail")
		requireNotificationTemplate(t, AdminAddTask.SubscribedInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
		requireEmailTemplates(t, "adminNotificationLinkerAddEmail")
		requireNotificationTemplate(t, AdminAddTask.AdminInternalNotificationTemplate(eventbus.TaskEvent{Outcome: eventbus.TaskOutcomeSuccess}))
	})
}
