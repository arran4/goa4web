package privateforum

import (
	"github.com/arran4/goa4web/core/templates"
	"testing"
)

func TestPrivateForumTasksTemplatesRequiredExist(t *testing.T) {
	tasks := []struct {
		name string
		task templatesRequired
	}{
		{"privateForumTask", &privateForumTask{}},
	}
	for _, task := range tasks {
		t.Run(task.name, func(t *testing.T) {
			req := task.task.TemplatesRequired()
			if len(req) == 0 {
				t.Fatalf("TemplatesRequired returned no templates; expected at least one")
			}
			for _, name := range req {
				if !templates.IsTemplateAvailable(name) {
					t.Fatalf("missing template: %s", name)
				}
			}
		})
	}
}

type templatesRequired interface {
	TemplatesRequired() []string
}
