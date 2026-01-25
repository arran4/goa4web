package auth

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/tasks"
)

type emailTemplatesRequired interface {
	EmailTemplatesRequired() []tasks.Page
}

func TestAuthTasksTemplatesExist(t *testing.T) {
	taskList := []struct {
		name string
		task interface{}
	}{
		{"LoginTask", &LoginTask{}},
		{"ForgotPasswordTask", &ForgotPasswordTask{}},
	}

	for _, entry := range taskList {
		t.Run(entry.name, func(t *testing.T) {
			if tr, ok := entry.task.(tasks.TemplatesRequired); ok {
				req := tr.TemplatesRequired()
				if len(req) == 0 {
					t.Errorf("TemplatesRequired returned no templates")
				}
				for _, p := range req {
					if !p.Exists() {
						t.Errorf("missing template: %s", p)
					}
				}
			}

			if etr, ok := entry.task.(emailTemplatesRequired); ok {
				req := etr.EmailTemplatesRequired()
				if len(req) == 0 {
					t.Errorf("EmailTemplatesRequired returned no templates")
				}
				for _, p := range req {
					// Emails can be in email/ or notifications/ depending on logic,
					// but NewEmailTemplates puts them in email/.
					// However, some might be shared?
					// Based on NewEmailTemplates ("email" based), they should be in email/
					if !templates.EmailTemplateExists(string(p)) {
						t.Errorf("missing email template: %s", p)
					}
				}
			}
		})
	}
}
