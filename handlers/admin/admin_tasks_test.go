package admin

import (
	"github.com/arran4/goa4web/core/templates"
	"testing"

	"github.com/arran4/goa4web/internal/tasks"
)

func TestHappyPathAdminTasksTemplatesRequiredExist(t *testing.T) {
	tasksList := []struct {
		name string
		task tasks.TemplatesRequired
	}{
		{"UserForcePasswordChangeTask", &UserForcePasswordChangeTask{}},
		{"UserSendResetEmailTask", &UserSendResetEmailTask{}},
		{"ServerShutdownTask", &ServerShutdownTask{}},
		// Note: ForgotPasswordTask is in `auth` package, so we can't test it directly here without importing `handlers/auth` which might cause cycle if not careful.
		// However, the user asked to ensure coverage. Ideally we should add a test in `handlers/auth` or move this helper.
		// For now, I will create a new test file in `handlers/auth` that mirrors this logic if I can't add it here.

	}
	for _, task := range tasksList {
		t.Run(task.name, func(t *testing.T) {
			req := task.task.RequiredTemplates()
			if len(req) == 0 {
				t.Fatalf("RequiredTemplates returned no templates; expected at least one")
			}
			for _, name := range req {
				if !name.Exists(templates.WithSilence(true)) {
					t.Fatalf("missing template: %s", name)
				}
			}
		})
	}
}
