package admin

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestAdminTasksTemplatesRequiredExist(t *testing.T) {
	tasks := []struct {
		name string
		task templatesRequired
	}{
		{"UserForcePasswordChangeTask", &UserForcePasswordChangeTask{}},
		{"UserSendResetEmailTask", &UserSendResetEmailTask{}},
		{"ServerShutdownTask", &ServerShutdownTask{}},
		// Note: ForgotPasswordTask is in `auth` package, so we can't test it directly here without importing `handlers/auth` which might cause cycle if not careful.
		// However, the user asked to ensure coverage. Ideally we should add a test in `handlers/auth` or move this helper.
		// For now, I will create a new test file in `handlers/auth` that mirrors this logic if I can't add it here.

	}
	for _, task := range tasks {
		t.Run(task.name, func(t *testing.T) {
			req := task.task.TemplatesRequired()
			if len(req) == 0 {
				t.Fatalf("TemplatesRequired returned no templates; expected at least one")
			}
			for _, name := range req {
				if !name.Exists() {
					t.Fatalf("missing template: %s", name)
				}
			}
			if etr, ok := task.task.(emailTemplatesRequired); ok {
				req := etr.EmailTemplatesRequired()
				if len(req) == 0 {
					t.Fatalf("EmailTemplatesRequired returned no templates; expected at least one")
				}
				for _, name := range req {
					if !templates.EmailTemplateExists(string(name)) {
						t.Fatalf("missing email template: %s", name)
					}
				}
			}
		})
	}
}

type emailTemplatesRequired interface {
	EmailTemplatesRequired() []tasks.Page
}

type templatesRequired interface {
	TemplatesRequired() []tasks.Page
}
