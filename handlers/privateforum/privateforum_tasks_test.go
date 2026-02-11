package privateforum

import (
	"github.com/arran4/goa4web/core/templates"
	"testing"

	"github.com/arran4/goa4web/internal/tasks"
)

func TestPrivateForumTasksTemplatesRequiredExist(t *testing.T) {
	tasks := []struct {
		name string
		task tasks.TemplatesRequired
	}{
		{"privateForumTask", &privateForumTask{}},
	}
	for _, task := range tasks {
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
