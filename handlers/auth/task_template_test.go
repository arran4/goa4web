package auth

import (
	"github.com/arran4/goa4web/core/templates"
	"testing"

	"github.com/arran4/goa4web/internal/tasks"
)

func TestAuthTasksTemplatesExist(t *testing.T) {
	taskList := []struct {
		name string
		task tasks.TemplatesRequired
	}{
		{"LoginTask", &LoginTask{}},
		{"ForgotPasswordTask", &ForgotPasswordTask{}},
	}

	for _, entry := range taskList {
		t.Run(entry.name, func(t *testing.T) {
			req := entry.task.RequiredTemplates()
			if len(req) == 0 {
				t.Errorf("RequiredTemplates returned no templates")
			}
			for _, p := range req {
				if !p.Exists(templates.WithSilence(true)) {
					t.Errorf("missing template: %s", p)
				}
			}
		})
	}
}
