package search

import (
	"testing"

	"github.com/arran4/goa4web/internal/tasks"
)

func TestSearchTasksTemplatesRequiredExist(t *testing.T) {
	tasks := []struct {
		name string
		task tasks.TemplatesRequired
	}{
		{"searchTask", &searchTask{}},
		{"RemakeBlogTask", &RemakeBlogTask{}},
		{"RemakeCommentsTask", &RemakeCommentsTask{}},
		{"RemakeImageTask", &RemakeImageTask{}},
		{"RemakeLinkerTask", &RemakeLinkerTask{}},
		{"RemakeNewsTask", &RemakeNewsTask{}},
		{"RemakeWritingTask", &RemakeWritingTask{}},
	}
	for _, task := range tasks {
		t.Run(task.name, func(t *testing.T) {
			req := task.task.RequiredTemplates()
			if len(req) == 0 {
				t.Fatalf("RequiredTemplates returned no templates; expected at least one")
			}
			for _, name := range req {
				if !name.Exists() {
					t.Fatalf("missing template: %s", name)
				}
			}
		})
	}
}
