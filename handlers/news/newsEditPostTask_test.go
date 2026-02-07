package news

import (
	"github.com/arran4/goa4web/core/templates"
	"testing"
)

func TestHappyPathEditTaskTemplatesRequiredExist(t *testing.T) {
	var task EditTask
	req := task.RequiredTemplates()
	if len(req) == 0 {
		t.Fatalf("EditTask.RequiredTemplates returned no templates; expected at least one")
	}
	for _, name := range req {
		if !name.Exists(templates.WithSilence(true)) {
			t.Fatalf("missing template: %s", name)
		}
	}
}
