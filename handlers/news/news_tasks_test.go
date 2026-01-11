package news

import (
	"github.com/arran4/goa4web/internal/tasks"
	"testing"
)

func TestNewsTasksTemplatesRequiredExist(t *testing.T) {
	tasks := []struct {
		name string
		task templatesRequired
	}{
		{"newsTask", &newsTask{}},
		{"newsPostTask", &newsPostTask{}},
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
