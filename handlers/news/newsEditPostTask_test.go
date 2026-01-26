package news

import (
	"testing"
)

func TestEditTaskTemplatesRequiredExist(t *testing.T) {
	var task EditTask
	req := task.RequiredTemplates()
	if len(req) == 0 {
		t.Fatalf("EditTask.RequiredTemplates returned no templates; expected at least one")
	}
	for _, name := range req {
		if !name.Exists() {
			t.Fatalf("missing template: %s", name)
		}
	}
}
