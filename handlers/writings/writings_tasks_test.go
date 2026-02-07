package writings

import (
	"github.com/arran4/goa4web/core/templates"
	"testing"

	"github.com/arran4/goa4web/internal/tasks"
)

func TestWritingsTasksTemplatesRequiredExist(t *testing.T) {
	t.Run("Happy Path - Templates Required", func(t *testing.T) {
		tsks := []struct {
			name string
			task tasks.TemplatesRequired
		}{
			{"writingsTask", &writingsTask{}},
		}
		for _, task := range tsks {
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
	})
}
