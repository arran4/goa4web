package news

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
)

func TestEditTaskTemplatesRequiredExist(t *testing.T) {
	var task EditTask
	req := task.TemplatesRequired()
	if len(req) == 0 {
		t.Fatalf("EditTask.TemplatesRequired returned no templates; expected at least one")
	}
	for _, name := range req {
		if !templates.TemplateExists(string(name)) {
			t.Fatalf("missing template: %s", name)
		}
	}
}
