package linker

import (
	"github.com/arran4/goa4web/internal/tasks"
	"testing"
)

func TestLinkerTasksTemplatesRequiredExist(t *testing.T) {
	tasks := []struct {
		name string
		task templatesRequired
	}{
		{"linkerTask", &linkerTask{}},
		{"linkerCategoryTask", &linkerCategoryTask{}},
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
