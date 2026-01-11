package admin

import (
	"github.com/arran4/goa4web/internal/tasks"
	"testing"
)

func TestAdminTasksTemplatesRequiredExist(t *testing.T) {
	tasks := []struct {
		name string
		task templatesRequired
	}{
		{"UserPasswordResetTask", &UserPasswordResetTask{}},
		{"ServerShutdownTask", &ServerShutdownTask{}},
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
		})
	}
}

type templatesRequired interface {
	TemplatesRequired() []tasks.Page
}
